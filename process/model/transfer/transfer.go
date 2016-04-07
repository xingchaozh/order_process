package transfer

import (
	"encoding/json"
	"github.com/Sirupsen/logrus"
	"order_process/process/db"
	"order_process/process/model/order"
)

// Transfer orders to current service
func Transfer(currentServiceId string, tranferredServiceId string, fn func(*order.OrderRecord)) error {
	return Reload(currentServiceId, tranferredServiceId, fn)
}

// Load orders to current service
func Reload(currentServiceId string, tranferredServiceId string, fn func(orderRecord *order.OrderRecord)) error {
	// Retrieve the orders from the transferred servive
	rawMaps, _ := db.Query("", order.OrderStateInServiceTable+":"+tranferredServiceId)

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
		if orderMap["order_state_in_service"].(string) == order.OSS_Active.String() {
			logrus.Debugf("Reload: [%v]", orderMap)
			record, err := order.Get(orderMap["order_id"].(string))
			if err != nil {
				return err
			}

			// Update the service order map
			if currentServiceId != tranferredServiceId {
				// Transaction should be introduced here
				order.UpdateOrderStateInService(tranferredServiceId, record.OrderID, order.OSS_Transferred.String())
				order.UpdateOrderStateInService(currentServiceId, record.OrderID, order.OSS_Active.String())
			}
			record.ServiceID = currentServiceId

			// Dispatch the the retrieved order to pipeline manager
			fn(record)
		}
	}
	return nil
}
