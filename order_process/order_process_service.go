package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"order_process/lib/consumer"
	"order_process/lib/handlers"
	"order_process/lib/model/order"
	"order_process/lib/model/pipeline"
	"order_process/lib/util"

	"github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
)

// The max count pipeline of per Order Processing Service Instance
var (
	MaxPipelineCount = 50
)

type NewTaskHandlerFunc func(string, pipeline.IPipeline) pipeline.ITaskHandler
type NewPipelineFunc func(func(string, pipeline.IPipeline) pipeline.ITaskHandler) pipeline.IPipeline

// Order Processing Service
type OrderProcessService struct {
	pipelines                 []pipeline.IPipeline
	lastPipelineSelectedIndex int
	ServiceID                 string
}

// The constructor of Order Processing Service
func NewOrderProcessService(NewPipeline NewPipelineFunc, NewTaskHandler NewTaskHandlerFunc) *OrderProcessService {
	service := OrderProcessService{
		lastPipelineSelectedIndex: -1,
		ServiceID:                 util.NewUUID(),
	}
	for i := 0; i < MaxPipelineCount; i++ {
		service.pipelines = append(service.pipelines, NewPipeline(NewTaskHandler))
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

	logrus.Debugf("POST Request Body: [%v]", t)

	// Generate order record
	t["user_id"] = tokenInfo.UserID
	t["service_id"] = this.ServiceID
	orderRecord, _ := order.New(t)
	logrus.Debugf("New order created with ID: [%v]", orderRecord.OrderID)

	// Create order processing job and process asynchronously using selected pipeline
	processJob := pipeline.NewProcessJob(orderRecord)
	this.SelectPipeline().AppendJob(processJob)

	// Generate response
	response := map[string]string{
		"order_id":   orderRecord.OrderID,
		"start_time": orderRecord.StartTime,
	}
	str, _ := json.Marshal(response)
	//str, _ := orderRecord.ToJson()
	w.Header().Add("Content-Type", "application/json")
	fmt.Fprint(w, string(str))
}

// GET /orders/{id}
func (this *OrderProcessService) QureyOrder(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	tokenInfo, err := consumer.GetTokenInfo(token)
	if err != nil || tokenInfo.UserID == "" {
		w.WriteHeader(401)
		return
	}

	id := mux.Vars(r)["id"]
	logrus.Debugf("Get Request ID: [%v]", id)

	record, err := order.Get(id)
	if err != nil {
		w.WriteHeader(404)
		return
	}

	str, _ := record.ToJson()

	w.Header().Add("Content-Type", "application/json")
	fmt.Fprint(w, tokenInfo.UserID, str)
}

// Round Robin Select pipeline
func (this *OrderProcessService) SelectPipeline() pipeline.IPipeline {
	if this.lastPipelineSelectedIndex+1 < len(this.pipelines) {
		this.lastPipelineSelectedIndex++
	} else {
		this.lastPipelineSelectedIndex = 0
	}
	return this.pipelines[this.lastPipelineSelectedIndex]
}

func (this *OrderProcessService) GetServiceID() string {
	return this.ServiceID
}

// Start Service
func (this *OrderProcessService) Start() (err error) {
	for _, pipeline := range this.pipelines {
		pipeline.Start()
	}

	r := mux.NewRouter()

	// Heartbeat
	r.HandleFunc("/diagnostic/heartbeat", handlers.NewHeartBeat(this.ServiceID).HeartBeatHandler).Methods("GET")

	// Create order
	r.HandleFunc("/orders", this.CreateOrder).Methods("POST")

	// Qurey specified order
	r.HandleFunc("/orders/{id}", this.QureyOrder).Methods("GET")

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
