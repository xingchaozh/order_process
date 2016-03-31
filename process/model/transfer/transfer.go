package transfer

import (
	"encoding/json"
	"order_process/process/db"
	"order_process/process/model/order"
	"order_process/process/model/pipeline"
)

// Transfer orders to current service
func Transfer(currentServiceId string, tranferredServiceId string, pipelineManager pipeline.IPipelineManager) error {
	// Retrieve the orders from the transferred servive
	rawMaps, _ := db.Query("", "ServiceOrderMap:"+tranferredServiceId)

	// Parse orders
	var ordersMap []map[string]interface{}
	for index, val := range rawMaps {
		if index%2 != 0 {
			t := make(map[string]interface{})
			err := json.Unmarshal([]byte(val[string(index)].(string)), &t)
			if err != nil {
				return err
			}
			ordersMap = append(ordersMap, t)
		}
	}

	// Loop the orders
	for _, orderMap := range ordersMap {
		if orderMap["order_state_in_service"].(string) == "Active" {
			record, err := order.Get(orderMap["order_id"].(string))
			if err != nil {
				return err
			}

			// Update the service order map
			// cluster.ServiceID
			order.UpdateOrderStateInService(tranferredServiceId, record.OrderID, "Transferred")
			order.UpdateOrderStateInService(currentServiceId, record.OrderID, "Active")
			record.ServiceID = currentServiceId

			// Dispatch the the retrieved order to pipeline manager
			pipelineManager.DispatchOrder(record)
		}
	}
	return nil
}
