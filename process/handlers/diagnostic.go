package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"order_process/process/model/cluster"
)

// The definition of heart beat check
type HeartBeat struct {
	ServiceID string
	Cluster   cluster.ICluster
}

// The constructor of heart beat check
func NewHeartBeat(serviceID string, cluster cluster.ICluster) *HeartBeat {
	return &HeartBeat{
		ServiceID: serviceID,
		Cluster:   cluster,
	}
}

// Heartbeat API handler, used for heartbeat check
func (this *HeartBeat) HeartBeatHandler(w http.ResponseWriter, r *http.Request) {
	// Generate response
	response := map[string]string{
		"serive_id": this.ServiceID,
		"status":    "OK",
	}

	str, _ := json.Marshal(response)
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, string(str))
}
