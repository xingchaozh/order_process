package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"order_process/process/consumer"
	"order_process/process/env"
	"order_process/process/handlers"
	"order_process/process/model/cluster"
	"order_process/process/model/order"
	"order_process/process/model/pipeline"
	"order_process/process/model/transfer"
	"order_process/process/util"
	"strconv"

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
	Addr            string
	Cluster         cluster.ICluster
}

// The constructor of Order Processing Service
func NewOrderProcessService(serviceCfg *env.ServiceCfg) *OrderProcessService {
	service := OrderProcessService{
		ServiceID: util.NewUUID(),
		PipelineManager: pipeline.NewProcessPipelineManager(MaxPipelineCount,
			pipeline.NewProcessPipeline, pipeline.NewStepTaskHandler),
		Addr: serviceCfg.IP + ":" + strconv.Itoa(serviceCfg.Port),
	}
	service.Cluster = cluster.New(service.ServiceID, service.Addr)
	return &service
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
		go transfer.Transfer(this.ServiceID, tranferredServiceId, this.PipelineManager)

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

// POST /diagnostic/register
func (this *OrderProcessService) RegisterService(w http.ResponseWriter, r *http.Request) {
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

	if _, ok := t["service_id"].(string); ok {
		this.Cluster.Register(t["service_id"].(string), t["service_addr"].(string))
	} else {
		w.WriteHeader(400)
		fmt.Fprint(w, "Invalid json format")
	}
}

func (this *OrderProcessService) RetrieveToken(r *http.Request) (*consumer.ConsumerInfo, error) {
	token := r.Header.Get("Authorization")
	tokenInfo, err := consumer.GetTokenInfo(token)
	if err != nil {
		return nil, err
	}
	return tokenInfo, nil
}

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

// Get current service id
func (this *OrderProcessService) GetServiceID() string {
	return this.ServiceID
}

// Start Service
func (this *OrderProcessService) Start() (err error) {
	// Start Cluster Management
	this.Cluster.Start()

	// Start pipeline
	this.PipelineManager.Start()

	r := mux.NewRouter()

	// Heartbeat
	r.HandleFunc("/diagnostic/heartbeat", handlers.NewHeartBeat(this.ServiceID, this.Cluster).HeartBeatHandler).Methods("GET")
	r.HandleFunc("/diagnostic/register", this.RegisterService).Methods("POST")

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
		Addr:    this.Addr,
		Handler: r,
	}
	logrus.Debugf("Service starts to listen on: [%v]", this.Addr)
	// Start Server
	return server.ListenAndServe()
}
