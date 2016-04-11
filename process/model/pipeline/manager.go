package pipeline

import (
	"order_process/process/model/order"
	"order_process/process/model/transfer"
)

// The interface of Pipleline Manager
type IPipelineManager interface {
	// Start the pipeline manager
	Start() error

	// Dispatch orders
	DispatchOrder(orderRecord *order.OrderRecord)

	// Stop the pipeline manager
	Stop()
}

// The definition of Order Process Pipeline Manager
type ProcessPipelineManager struct {
	pipelines                 []IPipeline
	lastPipelineSelectedIndex int
	serviceID                 string
}

// The constructor of Order Process Pipeline Manager
func NewProcessPipelineManager(serviceID string, MaxPipelineCount int,
	NewPipeline func(func(string, IPipeline) ITaskHandler) IPipeline,
	NewTaskHandler func(string, IPipeline) ITaskHandler) *ProcessPipelineManager {
	pipelineManager := ProcessPipelineManager{
		serviceID:                 serviceID,
		lastPipelineSelectedIndex: -1,
	}
	for i := 0; i < MaxPipelineCount; i++ {
		pipelineManager.pipelines = append(pipelineManager.pipelines, NewPipeline(NewTaskHandler))
	}
	return &pipelineManager
}

// Start the pipeline management and pipelines
func (this *ProcessPipelineManager) Start() error {
	for _, pipeline := range this.pipelines {
		pipeline.Start()
	}

	// Load the pending jobs
	fn := func(orderRecord *order.OrderRecord) {
		this.DispatchOrder(orderRecord)
	}
	return transfer.Reload(this.serviceID, this.serviceID, fn)
}

// Dispatch order assigned to pipeline manager
func (this *ProcessPipelineManager) DispatchOrder(orderRecord *order.OrderRecord) {
	processJob := NewProcessJob(orderRecord)
	this.SelectPipeline().AppendJob(processJob)
}

// Stop the pipeline management and pipelines
func (this *ProcessPipelineManager) Stop() {
	for _, pipeline := range this.pipelines {
		pipeline.Stop()
	}
}

// Round Robin Select pipeline
func (this *ProcessPipelineManager) SelectPipeline() IPipeline {
	if this.lastPipelineSelectedIndex+1 < len(this.pipelines) {
		this.lastPipelineSelectedIndex++
	} else {
		this.lastPipelineSelectedIndex = 0
	}
	return this.pipelines[this.lastPipelineSelectedIndex]
}
