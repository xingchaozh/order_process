package cluster

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/goraft/raft"
	"github.com/gorilla/mux"
)

// The interface of Cluster
type ICluster interface {
	Start(string)

	StateChangeEventHandler(raft.Event)
	LeaderChangeEventHandler(raft.Event)
	TermChangeEventHandler(raft.Event)

	RegisterService(io.ReadCloser) error
}

// The definition of Cluster
type Cluster struct {
	serviceID  string      `json:"service_id"`
	host       string      `json:"host"`
	port       int         `json:"port"`
	path       string      `json:"path"`
	router     *mux.Router `json:"mux_router"`
	raftServer raft.Server `json:"raft_server"`
	peers      map[string]bool
}

// The const used to check state of service
const (
	ClusterStatusCheckInterval = 10 // in seconds
	MaxHeartbeatFailTimes      = 5
)

// The constructor of Cluster
func New(serviceId string, host string, port int, path string, router *mux.Router) *Cluster {
	return &Cluster{
		serviceID: serviceId,
		host:      host,
		port:      port,
		path:      path,
		router:    router,
		peers:     make(map[string]bool),
	}
}

// Start the cluster
func (this *Cluster) Start(leader string) {
	var err error

	logrus.Printf("Initializing Raft Server: %s", this.path)

	// Initialize and start Raft server.
	transporter := raft.NewHTTPTransporter("/raft", 200*time.Millisecond)
	this.raftServer, err = raft.NewServer(this.serviceID, this.path, transporter, nil, nil, "")
	if err != nil {
		logrus.Fatal(err)
	}
	transporter.Install(this.raftServer, this)

	this.raftServer.AddEventListener(raft.StateChangeEventType, this.StateChangeEventHandler)
	this.raftServer.AddEventListener(raft.LeaderChangeEventType, this.LeaderChangeEventHandler)
	this.raftServer.AddEventListener(raft.TermChangeEventType, this.TermChangeEventHandler)

	this.raftServer.Start()

	// Join to the cluster
	if leader != "" {
		// Join to leader if specified.
		logrus.Println("Attempting to join leader:", leader)

		if !this.raftServer.IsLogEmpty() {
			logrus.Fatal("Cannot join with an existing log")
		}
		if err := this.Join(leader); err != nil {
			logrus.Fatal(err)
		}

	} else if this.raftServer.IsLogEmpty() {
		// Initialize the server by joining itself.
		logrus.Println("Initializing new cluster")

		_, err := this.raftServer.Do(&raft.DefaultJoinCommand{
			Name:             this.raftServer.Name(),
			ConnectionString: this.connectionString(),
		})
		if err != nil {
			logrus.Fatal(err)
		}

	} else {
		logrus.Println("Recovered from log")
	}

	// Discovery leader Service
	leaderName := this.raftServer.Leader()
	logrus.Println(leaderName)
	logrus.Println(this.raftServer.Peers())
}

// Returns the connection string.
func (this *Cluster) connectionString() string {
	return fmt.Sprintf("http://%s:%d", this.host, this.port)
}

// Joins to the leader of an existing cluster.
func (this *Cluster) Join(leader string) error {
	command := &raft.DefaultJoinCommand{
		Name:             this.raftServer.Name(),
		ConnectionString: this.connectionString(),
	}

	var b bytes.Buffer
	json.NewEncoder(&b).Encode(command)
	resp, err := http.Post(fmt.Sprintf("http://%s/cluster/join", leader), "application/json", &b)
	if err != nil {
		return err
	}
	resp.Body.Close()

	return nil
}

// Register Service
func (this *Cluster) RegisterService(body io.ReadCloser) error {
	defer body.Close()

	command := &raft.DefaultJoinCommand{}

	if err := json.NewDecoder(body).Decode(&command); err != nil {
		return err
	}
	if _, err := this.raftServer.Do(command); err != nil {
		return err
	}
	return nil
}

// This is a hack around Gorilla mux not providing the correct net/http HandleFunc() interface.
func (this *Cluster) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	this.router.HandleFunc(pattern, handler)
}

func (this *Cluster) StateChangeEventHandler(e raft.Event) {
	server := e.Source().(raft.Server)
	logrus.Printf("[%s] %s %v -> %v\n", server.Name(), e.Type(), e.PrevValue(), e.Value())
}

func (this *Cluster) LeaderChangeEventHandler(e raft.Event) {
	go this.LeaderChange(e)
}

func (this *Cluster) TermChangeEventHandler(e raft.Event) {
	server := e.Source().(raft.Server)
	logrus.Printf("[%s] %s %v -> %v\n", server.Name(), e.Type(), e.PrevValue(), e.Value())
}

func (this *Cluster) LeaderChange(e raft.Event) {
	server := e.Source().(raft.Server)
	logrus.Printf("[%s] %s %v -> %v", server.Name(), e.Type(), e.PrevValue(), e.Value())

	if this.IsCurrentServiceLeader() {
		logrus.Println("Start to perform leader tasks.")
		// Perform task
		this.CheckPeersStatus()
	}
}

// Check the services status, update the state if online, or set offline if failure.
func (this *Cluster) CheckPeersStatus() {
	ticker := time.NewTicker(time.Second * ClusterStatusCheckInterval)
	for _ = range ticker.C {
		if !this.IsCurrentServiceLeader() {
			return
		}

		logrus.Debugf("Check Peers Status: MemberCount [%v], Peers Count [%v]",
			this.raftServer.MemberCount(), len(this.raftServer.Peers()))

		for _, peer := range this.raftServer.Peers() {
			logrus.Debugf("Peer [%v]", peer)
			if this.IsPeerOffline(peer) {
				// Become OFFLINE
				if connected, ok := this.peers[peer.Name]; !ok || connected {
					this.peers[peer.Name] = false
					go this.TransferOrders(peer.Name)
				}
			} else {
				this.peers[peer.Name] = true
			}

			if !this.IsCurrentServiceLeader() {
				return
			}
		}
	}
}

func (this *Cluster) IsPeerOffline(peer *raft.Peer) bool {
	elapsedTime := time.Now().Sub(peer.LastActivity())
	if elapsedTime > time.Duration(float64(raft.DefaultHeartbeatInterval)*MaxHeartbeatFailTimes) {
		return true
	}
	return false
}

func (this *Cluster) IsCurrentServiceLeader() bool {
	return this.raftServer.State() == raft.Leader // this.raftServer.Name() == this.raftServer.Leader()
}

// Transfer the orders of one offline service
func (this *Cluster) TransferOrders(serviceId string) {
	transfer := func(connectionString string, client *http.Client) bool {
		data := map[string]string{
			"service_id": serviceId,
		}
		jsonData, _ := json.Marshal(data)
		body := strings.NewReader(string(jsonData))
		req, _ := http.NewRequest("POST", connectionString+"/service/transfer", body)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "user")
		resp, err := client.Do(req)
		if err == nil && resp.StatusCode == http.StatusOK {
			return true
		}
		return false
	}

	transferred := false
	client := &http.Client{}
	for !transferred && this.IsCurrentServiceLeader() {
		// Select one online service and transfer the pending orders
		for _, peer := range this.raftServer.Peers() {
			if !this.IsPeerOffline(peer) {
				transferred = transfer(peer.ConnectionString, client)
				if transferred {
					break
				}
			}
		}
		transferred = transfer(this.connectionString(), client)
		if transferred {
			break
		}
		time.Sleep(time.Second) // Retry later
	}
}
