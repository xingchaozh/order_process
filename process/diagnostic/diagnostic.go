package diagnostic

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Sirupsen/logrus"

	"order_process/process/model/cluster"
)

// The definition of heart beat check
type Diagnostic struct {
	serviceID string
	cluster   cluster.ICluster
}

// The constructor of heart beat check
func New(serviceID string, cluster cluster.ICluster) *Diagnostic {
	return &Diagnostic{
		serviceID: serviceID,
		cluster:   cluster,
	}
}

// Heartbeat API handler, used for heartbeat check
func (this *Diagnostic) HeartBeatHandler(w http.ResponseWriter, r *http.Request) {
	logrus.Debug("GET /diagnostic/heartbeat")

	// Generate response
	response := map[string]string{
		"serive_id": this.serviceID,
		"status":    "OK",
	}

	str, _ := json.Marshal(response)
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, string(str))
}

// Cluster State API handler, used for describe the cluster state
func (this *Diagnostic) ClusterStatusHandler(w http.ResponseWriter, req *http.Request) {
	logrus.Debug("GET /diagnostic/cluster")

	if this.cluster.IsCurrentServiceLeader() {
		str, _ := this.cluster.DescribeState()
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, string(str))
	} else {
		leader, err := this.cluster.GetLeaderConnectionString()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		logrus.Debugf("POST /diagnostic/cluster Redirect to %s", leader+"/diagnostic/cluster")
		http.Redirect(w, req, leader+"/diagnostic/cluster", http.StatusTemporaryRedirect)
		return
	}
}
