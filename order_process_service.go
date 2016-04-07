package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"

	"github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"

	"order_process/process/consumer"
	"order_process/process/env"
	"order_process/process/model/cluster"
	"order_process/process/model/order"
	"order_process/process/model/pipeline"
	"order_process/process/model/transfer"
	"order_process/process/util"
)

// The max count pipeline of per Order Processing Service Instance
var (
	MaxPipelineCount = 50
)

// Order Processing Service
type OrderProcessService struct {
	serviceID  string
	host       string
	port       int
	path       string
	router     *mux.Router
	httpServer *http.Server

	pipelineManager pipeline.IPipelineManager
	cluster         cluster.ICluster
}

// Creates a new server.
// The constructor of Order Processing Service
func NewOrderProcessService(serviceCfg *env.ServiceCfg) *OrderProcessService {
	s := OrderProcessService{
		host:   serviceCfg.IP,
		port:   serviceCfg.Port,
		path:   serviceCfg.Path,
		router: mux.NewRouter(),
	}

	// Read existing serviceID or generate a new one.
	if b, err := ioutil.ReadFile(filepath.Join(s.path, "service_id")); err == nil {
		s.serviceID = string(b)
	} else {
		s.serviceID = util.NewUUID() // fmt.Sprintf("%07x", rand.Int())[0:7]
		if err = ioutil.WriteFile(filepath.Join(s.path, "service_id"), []byte(s.serviceID), 0644); err != nil {
			panic(err)
		}
	}
	return &s
}

// Starts the Service.
func (this *OrderProcessService) Start(leader string) error {
	// Initialize and Start the Cluster Management
	this.cluster = cluster.New(this.serviceID, this.host, this.port, this.path, this.router)
	this.cluster.Start(leader)

	// Initialize and start pipeline
	this.pipelineManager = pipeline.NewProcessPipelineManager(MaxPipelineCount,
		pipeline.NewProcessPipeline, pipeline.NewStepTaskHandler)
	this.pipelineManager.Start()

	logrus.Println("Initializing HTTP server")

	// Initialize and start HTTP server.
	this.httpServer = &http.Server{
		Addr:    fmt.Sprintf("%s:%d", this.host, this.port),
		Handler: this.router,
	}

	// Join the cluster
	this.router.HandleFunc("/cluster/join", this.RegisterService).Methods("POST")

	// Create order
	this.router.HandleFunc("/orders", this.CreateOrder).Methods("POST")

	// Qurey specified order
	this.router.HandleFunc("/orders/{id}", this.QureyOrder).Methods("GET")

	// Transfer orders from specified service
	this.router.HandleFunc("/service/transfer", this.Transfer).Methods("POST")

	// Welcome infomation
	this.router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Welcome to Order Processing System!")
	}).Methods("GET")

	logrus.Println("Listening at:", fmt.Sprintf("%s:%d", this.host, this.port))

	return this.httpServer.ListenAndServe()

}

// POST /cluster/join
func (this *OrderProcessService) RegisterService(w http.ResponseWriter, req *http.Request) {
	logrus.Debug("POST /cluster/join")
	err := this.cluster.RegisterService(req.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// POST /orders
func (this *OrderProcessService) CreateOrder(w http.ResponseWriter, r *http.Request) {
	// Retrieve user information
	tokenInfo, err := this.RetrieveToken(r)
	if err != nil {
		w.WriteHeader(401)
		return
	}

	// Parse request body
	t, err := this.ParseRequestBody(r)
	if err != nil {
		fmt.Fprint(w, err)
		return
	}

	logrus.Debug("POST /orders")

	// Generate order record
	t["user_id"] = tokenInfo.UserID
	t["service_id"] = this.serviceID
	orderRecord, _ := order.New(t)
	logrus.Debugf("New order created with ID: [%v]", orderRecord.OrderID)

	// Create order processing job according to order record and
	// process asynchronously using selected pipeline by PipelineManager
	this.pipelineManager.DispatchOrder(orderRecord)

	// Generate response
	response := map[string]string{
		"order_id":   orderRecord.OrderID,
		"start_time": orderRecord.StartTime,
	}
	str, _ := json.Marshal(response)
	w.Header().Add("Content-Type", "application/json")
	fmt.Fprint(w, string(str))
}

// GET /orders/{order_id}
func (this *OrderProcessService) QureyOrder(w http.ResponseWriter, r *http.Request) {
	tokenInfo, err := this.RetrieveToken(r)
	if err != nil {
		w.WriteHeader(401)
		return
	}

	id := mux.Vars(r)["id"]
	logrus.Debugf("Get /orders/[%v] ", id)

	// Retrieve order record
	record, err := order.Get(id)
	if err != nil {
		w.WriteHeader(404)
		return
	}

	str, _ := record.ToJsonForUser()
	//str, _ := record.ToJson()
	w.Header().Add("Content-Type", "application/json")
	fmt.Fprint(w, tokenInfo.UserID, str)
}

// This API allows current service takes over the orders processing
// from some service which is down.
// POST /service/transfer
func (this *OrderProcessService) Transfer(w http.ResponseWriter, r *http.Request) {
	// Retrieve user information
	_, err := this.RetrieveToken(r)
	if err != nil {
		w.WriteHeader(401)
		return
	}

	// Parse request body
	t, err := this.ParseRequestBody(r)
	if err != nil {
		fmt.Fprint(w, err)
		return
	}

	logrus.Debugf("POST /service/transfer RequestBody: [%v]", t)

	if tranferredServiceId, ok := t["service_id"].(string); ok {
		// transfer the orders to current service
		go transfer.Transfer(this.serviceID, tranferredServiceId, this.pipelineManager)

		// Generate response
		response := map[string]string{
			"tranferred_service_id": tranferredServiceId,
			"current_service_id":    this.serviceID,
		}
		str, _ := json.Marshal(response)
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, string(str))
	} else {
		w.WriteHeader(400)
		fmt.Fprint(w, "Invalid json format")
	}
}

// Get token information
func (this *OrderProcessService) RetrieveToken(r *http.Request) (*consumer.ConsumerInfo, error) {
	token := r.Header.Get("Authorization")
	tokenInfo, err := consumer.GetTokenInfo(token)
	if err != nil {
		return nil, err
	}
	return tokenInfo, nil
}

// Parse the body information of request
func (this *OrderProcessService) ParseRequestBody(r *http.Request) (map[string]interface{}, error) {
	// Parse request body
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	t := make(map[string]interface{})
	err = json.Unmarshal([]byte(body), &t)
	if err != nil {
		return nil, err
	}
	return t, nil
}
