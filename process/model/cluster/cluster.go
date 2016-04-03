package cluster

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"order_process/process/db"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
)

// The interface of Cluster
type ICluster interface {
	Start()
	Register(serviceId string, addr string) error
	Offline(serviceId string, addr string) error
	DiscoveryCoordinator() (*ServiceInfomation, error)
}

// The definition of Cluster
type Cluster struct {
	CurrentServiceID string                        `json:"service_id"`
	CurrentAddr      string                        `json:"service_addr"`
	Services         map[string]*ServiceInfomation `json:"services"`
}

// The definition of Service Information
type ServiceInfomation struct {
	ServiceID     string       `json:"service_id"`
	Addr          string       `json:"service_addr"`
	State         ServiceState `json:"service_state"`
	UpdateTime    string       `json:"update_time"`
	RetryCount    int          `json:"retry_count"`
	LastCheckResp bool         `json:"last_check_responsed"`
}

// // The definition of Service State
type ServiceState int

const (
	ONLINE ServiceState = iota
	OFFLINE
	UNKNOWN
)

var ServiceStateName = map[ServiceState]string{
	ONLINE:  "Online",
	OFFLINE: "Offline",
	UNKNOWN: "Unknown",
}

func (s ServiceState) String() string {
	return ServiceStateName[s]
}

func ParseServiceState(state string) ServiceState {
	if state == ONLINE.String() {
		return ONLINE
	} else if state == OFFLINE.String() {
		return OFFLINE
	} else if state == UNKNOWN.String() {
		return UNKNOWN
	} else {
		return UNKNOWN
	}
}

// The definition of Tables
const (
	ServicesTable              = "Services"
	CoorServicesTable          = "CoordinatorServices"
	CoorServicesTableOnlineKey = "OnlineCoordinator"
)

// The const used to check state of service
const (
	HeartBeatInterval   = 10 // in seconds
	RetryCountIfFailure = 3
)

// The constructor of Cluster
func New(serviceId string, addr string) *Cluster {
	return &Cluster{
		CurrentServiceID: serviceId,
		CurrentAddr:      addr,
		Services:         make(map[string]*ServiceInfomation),
	}
}

// Start the cluster
func (this *Cluster) Start() {
	// Discovery Coordinator Service
	coorService, _ := this.DiscoveryCoordinator()
	if coorService != nil {
		logrus.Debug(coorService)
		retryCount := 0
		for retryCount <= RetryCountIfFailure {
			resp, err := http.Get("http://" + coorService.Addr + "/diagnostic/heartbeat")
			if err != nil {
				logrus.Error(err)
			} else if resp.StatusCode == http.StatusOK {
				err = this.RegisterCurrentToCoordinator(coorService)
				if err == nil {
					// Follower, Do nothing for the moment
					return
				}
				logrus.Error(err)
			}
			retryCount++
		}
	}

	// No Coordinator found
	logrus.Info("No Coordinator found")

	go this.PerformCoordinatorTask()
}

// Perform the tasks as a coordinator
func (this *Cluster) PerformCoordinatorTask() {
	// Register service
	this.Register(this.CurrentServiceID, this.CurrentAddr)

	// Select the Coordinator
	// TODO: Use raft algorithm to vote the leader
	this.SetCurrentAsCoordinator() // To simple the design, use current as leader

	// Retrieve Services information
	this.RetrieveAllServices()

	// Start the task of Coordinator
	this.CheckServicesStatus()
}

// Update service state to database
func UpdateServiceState(serviceId string, addr string, state ServiceState) error {
	// Generate service information
	service := GenerateServiceInfo(serviceId, addr, state.String(), time.Now().String())
	return db.Write(service.ToJson(), ServicesTable, serviceId)
}

// Register one service and online
func (this *Cluster) Register(serviceId string, addr string) error {
	this.Services[serviceId] = GenerateServiceInfo(serviceId, addr, ONLINE.String(), time.Now().String())
	return UpdateServiceState(serviceId, addr, ONLINE)
}

// Offline one service
func (this *Cluster) Offline(serviceId string, addr string) error {
	delete(this.Services, serviceId)
	return UpdateServiceState(serviceId, addr, OFFLINE)
}

// Discovery the coordinator service
func (this *Cluster) DiscoveryCoordinator() (*ServiceInfomation, error) {
	recordMap := make(map[string]interface{})
	err := db.Read("", recordMap, CoorServicesTable, CoorServicesTableOnlineKey)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}
	serviceMap := make(map[string]interface{})
	err = json.Unmarshal(recordMap[CoorServicesTableOnlineKey].([]byte), &serviceMap)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}
	if serviceMap["service_state"].(string) == ONLINE.String() {
		coorService := GenerateServiceInfo(serviceMap["service_id"].(string),
			serviceMap["service_addr"].(string),
			serviceMap["service_state"].(string),
			serviceMap["update_time"].(string))

		return coorService, nil
	}
	err = errors.New("No Coordinator Serice found")
	logrus.Error(err)
	return nil, err
}

// Get the services registered
func (this *Cluster) RetrieveAllServices() error {
	rawMaps, _ := db.Query("", ServicesTable)

	// Parse services
	var servicesMap []map[string]interface{}
	for index, val := range rawMaps {
		if index%2 != 0 {
			t := make(map[string]interface{})
			err := json.Unmarshal([]byte(val[string(index)].(string)), &t)
			if err != nil {
				return err
			}
			servicesMap = append(servicesMap, t)
		}
	}

	// Loop the services
	for _, serviceMap := range servicesMap {
		if serviceMap["service_state"].(string) == ONLINE.String() {
			service := GenerateServiceInfo(serviceMap["service_id"].(string),
				serviceMap["service_addr"].(string),
				serviceMap["service_state"].(string),
				serviceMap["update_time"].(string))
			this.Services[service.ServiceID] = service
		}
	}
	return nil
}

// Set current service as the coordinator
func (this *Cluster) SetCurrentAsCoordinator() {
	serviceInfo := GenerateServiceInfo(this.CurrentServiceID,
		this.CurrentAddr, ONLINE.String(), time.Now().String())
	db.Write(serviceInfo.ToJson(), CoorServicesTable, CoorServicesTableOnlineKey)
}

// Check the services status, update the state if online, or set offline if failure.
func (this *Cluster) CheckServicesStatus() {
	ticker := time.NewTicker(time.Second * HeartBeatInterval)
	for _ = range ticker.C {
		for _, serviceInfo := range this.Services {
			if serviceInfo.LastCheckResp && serviceInfo.ServiceID != this.CurrentServiceID {
				this.Services[serviceInfo.ServiceID].LastCheckResp = false
				go this.CheckSpecifiedServiceStatus(serviceInfo)
			}
		}
	}
}

// Check the specified service status
func (this *Cluster) CheckSpecifiedServiceStatus(serviceInfo *ServiceInfomation) {
	resp, err := http.Get("http://" + serviceInfo.Addr + "/diagnostic/heartbeat")
	if err != nil {
		logrus.Errorf("Cluster::CheckServicesStatus error [%v]", err)
	}

	if err != nil || resp.StatusCode != http.StatusOK {
		logrus.Debugf("Not reached: [%v]", serviceInfo)

		this.Services[serviceInfo.ServiceID].RetryCount++
		if this.Services[serviceInfo.ServiceID].RetryCount >= RetryCountIfFailure {
			this.Offline(serviceInfo.ServiceID, serviceInfo.Addr)

			logrus.Debugf("Offline: [%v]", serviceInfo)
			go this.TransferOrders(serviceInfo)
		}
	} else {
		//*
		defer resp.Body.Close()
		data, _ := ioutil.ReadAll(resp.Body)
		logrus.Debugf("Reached: [%v][%v]", serviceInfo, string(data))
		logrus.Debugln()
		//*/

		this.Services[serviceInfo.ServiceID].UpdateTime = time.Now().String()
		this.Services[serviceInfo.ServiceID].RetryCount = 0
		UpdateServiceState(serviceInfo.ServiceID, serviceInfo.Addr, ONLINE)
	}

	if _, found := this.Services[serviceInfo.ServiceID]; found {
		this.Services[serviceInfo.ServiceID].LastCheckResp = true
	}
}

// Register current service to the coordinator
func (this *Cluster) RegisterCurrentToCoordinator(coorService *ServiceInfomation) error {
	client := &http.Client{}
	data := map[string]string{
		"service_id":   this.CurrentServiceID,
		"service_addr": this.CurrentAddr,
	}
	jsonData, _ := json.Marshal(data)
	body := strings.NewReader(string(jsonData))
	req, _ := http.NewRequest("POST", "http://"+coorService.Addr+"/diagnostic/register", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "user")
	resp, err := client.Do(req)
	if err != nil {
		return err
	} else if resp.StatusCode == http.StatusOK {
		return nil
	}
	return errors.New("Register Current To Coordinator Failed")
}

// Transfer the orders of one offline service
func (this *Cluster) TransferOrders(serviceInfo *ServiceInfomation) {
	transferred := false
	client := &http.Client{}
	for !transferred {
		// Select one online service and transfer the pending orders
		for _, service := range this.Services {
			if service.State == ONLINE && service.RetryCount == 0 {
				data := map[string]string{
					"service_id": serviceInfo.ServiceID,
				}
				jsonData, _ := json.Marshal(data)
				body := strings.NewReader(string(jsonData))
				req, _ := http.NewRequest("POST", "http://"+service.Addr+"/service/transfer", body)
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "user")
				resp, err := client.Do(req)
				if err == nil && resp.StatusCode == http.StatusOK {
					transferred = true
					break
				}
				time.Sleep(time.Second) // Retry later
			}
		}
	}
}

// To map format
func (this *ServiceInfomation) ToMap() map[string]string {
	// Generate service information
	return map[string]string{
		"service_id":    this.ServiceID,
		"service_addr":  this.Addr,
		"service_state": this.State.String(),
		"update_time":   this.UpdateTime,
	}
}

// To json file format
func (this *ServiceInfomation) ToJson() string {
	jsonInfo, _ := json.Marshal(this.ToMap())
	return string(jsonInfo)
}

func GenerateServiceInfo(serviceId string, addr string, state string, updateTime string) *ServiceInfomation {
	return &ServiceInfomation{
		ServiceID:     serviceId,
		Addr:          addr,
		State:         ParseServiceState(state),
		UpdateTime:    updateTime,
		RetryCount:    0,
		LastCheckResp: true,
	}
}
