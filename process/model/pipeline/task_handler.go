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
	AppendTask(job IJob) error

	// Start perform tasks
	PerformTasks()

	// Rollback current task if failure
	Rollback() error

	// Stop the task handler
	Stop()
}

// The dedicated step task handler
type ProcessStepTaskHandler struct {
	StepTaskType   string
	PendingTasks   chan IJob
	CurentStepTask IJob
	PipeLine       IPipeline
	stopped        bool
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
		stopped:      false,
	}
}

// Append task to pending list
func (this *ProcessStepTaskHandler) AppendTask(job IJob) error {
	if !this.stopped {
		this.PendingTasks <- job.(*ProcessJob)
		return nil
	}
	return errors.New("The target task handler has been stopped.")
}

// Loop the pending list and process
func (this *ProcessStepTaskHandler) PerformTasks() {
	for this.CurentStepTask = range this.PendingTasks {
		this.HandleCurrentTask()

		if this.stopped {
			break
		}
	}
}

// Handle the task
func (this *ProcessStepTaskHandler) HandleCurrentTask() error {
	logrus.Debugf("[%s]handling step[%s]", this.CurentStepTask.GetJobID(), this.StepTaskType)

	var err error

	if this.CurentStepTask.IsJobRollbacking() && this.StepTaskType != "Failed" {
		err = this.Rollback()
	} else {
		err = this.StartStep()
		if err == nil {
			// Simulate the processing of current order step
			if !this.CurentStepTask.IsJobInFinishingStep() {
				time.Sleep(time.Second * StepProcessTime)
			}

			// Simulate the processing failure occurs at 5% ratio
			if util.IsEventWithSpecifiedRatioHappens() && !this.CurentStepTask.IsJobInFinishingStep() {
				err = errors.New("Random failure")
			} else {
				err = this.FinishStep()
			}
		}

	}

	if err != nil {
		this.CurentStepTask.MarkJobAsFailure()
		// Trigger roll back
		this.CurentStepTask.StartRollback()

		logrus.Debugf("[%s]Failure occurs when handling step[%s][%v]",
			this.CurentStepTask.GetJobID(), this.CurentStepTask.GetCurrentStep(), err)
	}

	go this.PipeLine.DispatchTask(this.CurentStepTask.GetJobID())
	return err
}

// Handle the rollback operation
func (this *ProcessStepTaskHandler) Rollback() error {
	logrus.Debugf("[%s]Rollback step[%s]", this.CurentStepTask.GetJobID(), this.StepTaskType)

	return this.CurentStepTask.RollbackStep(this.StepTaskType)
}

// Start current step
func (this *ProcessStepTaskHandler) StartStep() error {
	job := this.CurentStepTask
	err := VerifyStepSwitch(job.GetCurrentStep(), this.StepTaskType)
	if err != nil {
		return err
	}

	err = job.StartStep(this.StepTaskType)
	logrus.Debugf("[%s]Start step[%s]", job.GetJobID(), job.GetCurrentStep())
	return err
}

// Finish current step
func (this *ProcessStepTaskHandler) FinishStep() error {
	job := this.CurentStepTask
	if this.StepTaskType != job.GetCurrentStep() {
		return errors.New("Cannot finish step since it is not current step")
	}

	err := job.FinishCurrentStep()
	logrus.Debugf("[%s]Finish step[%s]", job.GetJobID(), job.GetCurrentStep())
	return err
}

// Stop the task handler
func (this *ProcessStepTaskHandler) Stop() {
	if !this.stopped {
		this.stopped = true
		close(this.PendingTasks)
	}
}
