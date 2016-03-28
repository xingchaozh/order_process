package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"order_process/process/consumer"
	"order_process/process/handlers"
	"order_process/process/model/order"
	"order_process/process/model/pipeline"
	"order_process/process/model/transfer"
	"order_process/process/util"

	"github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
)

// The max count pipeline of per Order Processing Service Instance
var (
	MaxPipelineCount = 50
)

// Order Processing Service
type OrderProcessService struct {
	ServiceID       string
	PipelineManager pipeline.IPipelineManager
}

// The constructor of Order Processing Service
func NewOrderProcessService() *OrderProcessService {
	service := OrderProcessService{
		ServiceID: util.NewUUID(),
		PipelineManager: pipeline.NewProcessPipelineManager(MaxPipelineCount,
			pipeline.NewProcessPipeline, pipeline.NewStepTaskHandler),
	}
	return &service
}

// POST /orders
func (this *OrderProcessService) CreateOrder(w http.ResponseWriter, r *http.Request) {
	// Retrieve user information
	token := r.Header.Get("Authorization")
	tokenInfo, err := consumer.GetTokenInfo(token)
	if err != nil || tokenInfo.UserID == "" {
		w.WriteHeader(401)
		return
	}

	// Parse request body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprint(w, err)
		return
	}
	t := make(map[string]interface{})
	err = json.Unmarshal([]byte(body), &t)
	if err != nil {
		fmt.Fprint(w, err)
		return
	}

	logrus.Debugf("POST /orders RequestBody: [%v]", t)

	// Generate order record
	t["user_id"] = tokenInfo.UserID
	t["service_id"] = this.ServiceID
	orderRecord, _ := order.New(t)
	logrus.Debugf("New order created with ID: [%v]", orderRecord.OrderID)

	// Create order processing job according to order record and
	// process asynchronously using selected pipeline by PipelineManager
	this.PipelineManager.DispatchOrder(orderRecord)

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
	token := r.Header.Get("Authorization")
	tokenInfo, err := consumer.GetTokenInfo(token)
	if err != nil || tokenInfo.UserID == "" {
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
	token := r.Header.Get("Authorization")
	tokenInfo, err := consumer.GetTokenInfo(token)
	if err != nil || tokenInfo.UserID == "" {
		w.WriteHeader(401)
		return
	}

	// Parse request body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprint(w, err)
		return
	}
	t := make(map[string]interface{})
	err = json.Unmarshal([]byte(body), &t)
	if err != nil {
		fmt.Fprint(w, err)
		return
	}

	logrus.Debugf("POST /service/transfer RequestBody: [%v]", t)

	if tranferredServiceId, ok := t["service_id"].(string); ok {
		// transfer the orders to current service
		go transfer.Transfer(tranferredServiceId, this.PipelineManager)

		// Generate response
		response := map[string]string{
			"tranferred_service_id": tranferredServiceId,
			"current_service_id":    this.ServiceID,
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

// Get current service id
func (this *OrderProcessService) GetServiceID() string {
	return this.ServiceID
}

// Start Service
func (this *OrderProcessService) Start() (err error) {
	this.PipelineManager.Start()

	r := mux.NewRouter()

	// Heartbeat
	r.HandleFunc("/diagnostic/heartbeat", handlers.NewHeartBeat(this.ServiceID).HeartBeatHandler).Methods("GET")

	// Create order
	r.HandleFunc("/orders", this.CreateOrder).Methods("POST")

	// Qurey specified order
	r.HandleFunc("/orders/{id}", this.QureyOrder).Methods("GET")

	// Transfer orders from specified service
	r.HandleFunc("/service/transfer", this.Transfer).Methods("POST")

	// Welcome infomation
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Welcome to Order Processing System!")
	}).Methods("GET")

	// Initilize Server
	server := http.Server{
		Addr:    "127.0.0.1:8080",
		Handler: r,
	}
	// Start Server
	return server.ListenAndServe()
}
