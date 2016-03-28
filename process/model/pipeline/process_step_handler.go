package pipeline

import (
	"errors"
	"order_process/process/util"
	"time"

	"github.com/Sirupsen/logrus"
)

// The interface of task handler for Order Step Processing
type ITaskHandler interface {
	// Append new task
	AppendTask(job IJob)

	// Start perform tasks
	PerformTasks()

	// Rollback current task if failure
	Rollback()
}

// The dedicated step task handler
type ProcessStepTaskHandler struct {
	StepTaskType   string
	PendingTasks   chan IJob
	CurentStepTask IJob
	PipeLine       IPipeline
}

const (
	MaxPendingTasksCount = 10000
	StepProcessTime      = 5 //seconds
)

// The constructor of task handler for Order Step Processing
func NewStepTaskHandler(stepTaskType string, pipeLine IPipeline) ITaskHandler {
	return &ProcessStepTaskHandler{
		StepTaskType: stepTaskType,
		PendingTasks: make(chan IJob, MaxPendingTasksCount),
		PipeLine:     pipeLine,
	}
}

// Append task to pending list
func (this *ProcessStepTaskHandler) AppendTask(job IJob) {
	this.PendingTasks <- job.(*ProcessJob)
}

// Loop the pending list and process
func (this *ProcessStepTaskHandler) PerformTasks() {
	for this.CurentStepTask = range this.PendingTasks {
		this.HandleCurrentTask()
	}
}

// Handle the task
func (this *ProcessStepTaskHandler) HandleCurrentTask() error {
	logrus.Debugf("[%s]handling step[%s]", this.CurentStepTask.GetJobID(), this.StepTaskType)

	if this.CurentStepTask.IsJobRollbacking() && this.StepTaskType != "Failed" {
		this.Rollback()
	} else {
		this.StartStep()
		// Simulate the processing of current order step
		if !this.CurentStepTask.IsJobInFinishingStep() {
			time.Sleep(time.Second * StepProcessTime)
		}

		// Simulate the processing failure occurs at 5% ratio
		if util.IsEventWithSpecifiedRatioHappens() && !this.CurentStepTask.IsJobInFinishingStep() {
			this.CurentStepTask.MarkJobAsFailure()

			// Trigger roll back
			this.CurentStepTask.StartRollback()

			logrus.Debugf("[%s]Error occurs when handling step[%s]",
				this.CurentStepTask.GetJobID(), this.CurentStepTask.GetCurrentStep())
		} else {
			this.FinishStep()
		}
	}

	go this.PipeLine.DispatchTask(this.CurentStepTask.GetJobID())
	return nil
}

// Handle the rollback operation
func (this *ProcessStepTaskHandler) Rollback() {
	logrus.Debugf("[%s]Rollback step[%s]", this.CurentStepTask.GetJobID(), this.StepTaskType)

	this.CurentStepTask.RollbackStep(this.StepTaskType)
}

// Start current step
func (this *ProcessStepTaskHandler) StartStep() error {
	job := this.CurentStepTask
	err := VerifyStepSwitch(job.GetCurrentStep(), this.StepTaskType)
	if err != nil {
		return err
	}

	job.StartStep(this.StepTaskType)

	logrus.Debugf("[%s]Start step[%s]", job.GetJobID(), job.GetCurrentStep())
	return nil
}

// Finish current step
func (this *ProcessStepTaskHandler) FinishStep() error {
	job := this.CurentStepTask
	if this.StepTaskType != job.GetCurrentStep() {
		return errors.New("Cannot finish step since it is not current step")
	}

	job.FinishCurrentStep()

	logrus.Debugf("[%s]Finish step[%s]", job.GetJobID(), job.GetCurrentStep())
	return nil
}
