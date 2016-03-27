package order

import (
	"encoding/json"
	"errors"
	"order_process/lib/db"
	"order_process/lib/util"
	"time"
)

type OrderStep struct {
	StepName       string `json:"step_name"`
	StartTime      string `json:"step_start_time"`
	CompleteTime   string `json:"step_complete_time"`
	StepCompleted  bool   `json:"step_completed"`
	StepRollbacked bool   `json:"step_rollbacked"`
}

type OrderRecord struct {
	OrderID        string      `json:"order_id"`
	CurrentStep    string      `json:"current_step"`
	StartTime      string      `json:"start_time"`
	CompleteTime   string      `json:"complete_time"`
	Steps          []OrderStep `json:"steps"`
	UserID         string      `json:"user_id"`
	Finished       bool        `json:"finished"`
	FailureOccured bool        `json:"failure_occured"`
	ServiceID      string      `json:"service_id"`
	RollbackState  string      `json:"rollback_state"`
}

// New order record
func New(record map[string]interface{}) (*OrderRecord, error) {
	id := util.NewUUID()
	record["order_id"] = id
	record["current_step"] = "Scheduling"
	record["start_time"] = time.Now().UTC().String()

	steps := []OrderStep{}
	orderStep := OrderStep{
		StepName:       record["current_step"].(string),
		StartTime:      time.Now().UTC().String(),
		StepCompleted:  false,
		StepRollbacked: false,
	}
	steps = append(steps, orderStep)
	record["steps"] = steps
	record["finished"] = false
	record["failure_occured"] = false
	record["rollback_state"] = "NA"

	orderRecord, err := GenerateOrderRecord(record)
	if err != nil {
		return nil, err
	}

	str, err := orderRecord.ToJson()
	err = db.Write(str, nil, "Order", id)
	if err != nil {
		return nil, err
	}

	return orderRecord, nil
}

// Generate Order Record according information stored in record map
func GenerateOrderRecord(record map[string]interface{}) (*OrderRecord, error) {
	if record["order_id"] == nil {
		return nil, errors.New("order_id is required")
	}
	if record["current_step"] == nil {
		return nil, errors.New("current_step is required")
	}
	if record["start_time"] == nil {
		return nil, errors.New("start_time is required")
	}
	if record["steps"] == nil {
		return nil, errors.New("steps is required")
	}
	if record["user_id"] == nil {
		return nil, errors.New("user_id is required")
	}

	parseStep := func(stepMap map[string]interface{}) OrderStep {
		step := OrderStep{
			StepName:       stepMap["step_name"].(string),
			StartTime:      stepMap["step_start_time"].(string),
			StepCompleted:  stepMap["step_completed"].(bool),
			StepRollbacked: stepMap["step_rollbacked"].(bool),
		}
		if v, ok := stepMap["step_complete_time"].(string); ok {
			step.CompleteTime = v
		}
		return step
	}

	var steps []OrderStep
	if stepsMaps, ok := record["steps"].([]string); ok {
		for _, jsonStep := range stepsMaps {
			stepMap := map[string]interface{}{}
			err := json.Unmarshal([]byte(jsonStep), &stepMap)
			if err != nil {
				return nil, err
			}
			steps = append(steps, parseStep(stepMap))
		}
	} else if stepsMaps, ok := record["steps"].([]map[string]interface{}); ok {
		for _, stepMap := range stepsMaps {
			steps = append(steps, parseStep(stepMap))
		}
	} else if _, ok := record["steps"].([]OrderStep); ok {
		steps, _ = record["steps"].([]OrderStep)
	} else {
		for _, stepMapInterface := range record["steps"].([]interface{}) {
			stepMap := stepMapInterface.(map[string]interface{})
			steps = append(steps, parseStep(stepMap))
		}
	}

	orderRecord := OrderRecord{
		OrderID:        record["order_id"].(string),
		CurrentStep:    record["current_step"].(string),
		StartTime:      record["start_time"].(string),
		Steps:          steps,
		UserID:         record["user_id"].(string),
		Finished:       record["finished"].(bool),
		FailureOccured: record["failure_occured"].(bool),
		ServiceID:      record["service_id"].(string),
		RollbackState:  record["rollback_state"].(string),
	}
	if orderRecord.Finished {
		orderRecord.CompleteTime = record["complete_time"].(string)
	}
	return &orderRecord, nil
}

// To json file
func (this *OrderRecord) ToJson() (string, error) {
	jsonOrderRecord, err := json.Marshal(this.ToMap())
	if err != nil {
		return "", err
	}
	return string(jsonOrderRecord), nil
}

func (this *OrderRecord) String() string {
	str, _ := this.ToJson()
	return str
}

// To map
func (this *OrderRecord) ToMap() *map[string]interface{} {
	stepsMap := []map[string]interface{}{}
	for _, step := range this.Steps {
		stepMap := map[string]interface{}{
			"step_name":       step.StepName,
			"step_start_time": step.StartTime,
			"step_completed":  step.StepCompleted,
			"step_rollbacked": step.StepRollbacked,
		}
		if step.StepCompleted {
			stepMap["step_complete_time"] = step.CompleteTime
		}
		stepsMap = append(stepsMap, stepMap)
	}

	recordMap := map[string]interface{}{
		"order_id":        this.OrderID,
		"current_step":    this.CurrentStep,
		"start_time":      this.StartTime,
		"steps":           stepsMap,
		"user_id":         this.UserID,
		"finished":        this.Finished,
		"failure_occured": this.FailureOccured,
		"service_id":      this.ServiceID,
		"rollback_state":  this.RollbackState,
	}

	if this.Finished {
		recordMap["complete_time"] = this.CompleteTime
	}
	return &recordMap
}

// Just used for query
func (this *OrderRecord) ToMapForQuery() *map[string]interface{} {
	stepsMap := []map[string]interface{}{}
	for _, step := range this.Steps {
		stepMap := map[string]interface{}{
			"step_name":       step.StepName,
			"step_start_time": step.StartTime,
		}
		if step.StepCompleted {
			stepMap["step_complete_time"] = step.CompleteTime
		}
		stepsMap = append(stepsMap, stepMap)
	}

	recordMap := map[string]interface{}{
		"order_id":     this.OrderID,
		"current_step": this.CurrentStep,
		"start_time":   this.StartTime,
		"steps":        stepsMap,
		"finished":     this.Finished,
	}

	if this.Finished {
		recordMap["complete_time"] = this.CompleteTime
	}
	return &recordMap
}

// Read from Database
func Get(orderId string) (*OrderRecord, error) {
	err := util.ValidateUUID(orderId)
	if err != nil {
		return nil, err
	}
	recordMap := make(map[string]interface{})
	db.Read("", &recordMap, "Order", orderId)

	t := make(map[string]interface{})
	err = json.Unmarshal(recordMap[orderId].([]byte), &t)
	if err != nil {
		return nil, err
	}
	orderRecord, err := GenerateOrderRecord(t)
	if err != nil {
		return nil, err
	}
	return orderRecord, nil
}
