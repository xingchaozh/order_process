package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type HeartBeat struct {
	ServiceID string
}

func NewHeartBeat(serviceID string) *HeartBeat {
	return &HeartBeat{
		ServiceID: serviceID,
	}
}

// Heartbeat, used for heartbeat check
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
