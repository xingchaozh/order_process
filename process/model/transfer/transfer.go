package transfer

import (
	"order_process/process/db"
	"order_process/process/model/order"
	"order_process/process/model/pipeline"
)

func Transfer(ServiceId string, pipelineManager pipeline.IPipelineManager) error {
	rawMaps, _ := db.Query("", "ServiceOrderMap:"+ServiceId)
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

	for _, orderIdMap := range orderIdsMap {
		if orderIdMap["orderState"].(string) == "Active" {
			record, err := order.Get(orderIdMap["orderId"].(string))
			if err != nil {
				return err
			}

			pipelineManager.DispatchOrder(record)
		}
	}
	return nil
}
