package transfer

import (
	"order_process/process/db"
	"order_process/process/model/order"
	"order_process/process/model/pipeline"
)

// Transfer orders to current service
func Transfer(ServiceId string, pipelineManager pipeline.IPipelineManager) error {
	// Retrieve the orders from the transferred servive
	rawMaps, _ := db.Query("", "ServiceOrderMap:"+ServiceId)

	// Parse orders
	var orderIdsMap []map[string]interface{}
	for index, val := range rawMaps {
		if index%2 == 0 {
			record := map[string]interface{}{
				"orderId": val[string(index)].(string),
			}
			orderIdsMap = append(orderIdsMap, record)
		} else {
			orderIdsMap[len(orderIdsMap)-1]["orderState"] = val[string(index)].(string)
		}
	}

	// Loop the orders
	for _, orderIdMap := range orderIdsMap {
		if orderIdMap["orderState"].(string) == "Active" {
			record, err := order.Get(orderIdMap["orderId"].(string))
			if err != nil {
				return err
			}
			// Dispatch the the retrieved order to pipeline manager
			pipelineManager.DispatchOrder(record)
		}
	}
	return nil
}
