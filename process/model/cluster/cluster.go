package cluster

import (
	"encoding/json"
	"order_process/process/db"
	"time"
)

type ICluster interface {
	Register(serviceId string) error
	Offline(serviceId string) error
	DiscoveryCoordinator()
}

type Cluster struct {
	ServiceID string
}

func New(serviceId string) *Cluster {
	return &Cluster{
		ServiceID: serviceId,
	}
}

func UpdateServiceState(serviceId string, state string) error {
	// Generate service information
	regInfo := map[string]string{
		"service_id":    serviceId,
		"service_state": state,
		"update_time":   time.Now().String(),
	}
	regInfoJson, _ := json.Marshal(regInfo)
	return db.Write(string(regInfoJson), "Services", serviceId)
}

func (this *Cluster) Register(serviceId string) error {
	return UpdateServiceState(serviceId, "ONLINE")
}

func (this *Cluster) Offline(serviceId string) error {
	return UpdateServiceState(serviceId, "OFFLINE")
}

func (this *Cluster) DiscoveryCoordinator() {

}
