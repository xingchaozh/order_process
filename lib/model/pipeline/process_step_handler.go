package pipeline

import (
	"errors"
	"order_process/lib/util"
	"time"

	"github.com/Sirupsen/logrus"
)

// The interface of task handler for Order Step Processing
type ITaskHandler interface {
	// Start perform tasks
	PerformTasks()
	// Append new task
	AppendTask(job IJob)
	// Rollback current task if failure
	Rollback()
}

type message struct {
	r      map[string]interface{}
	status string
	err    error
}

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

func (this *ProcessStepTaskHandler) PerformTasks() {
	for this.CurentStepTask = range this.PendingTasks {
		this.HandleCurrentTask()
	}
}

func (this *ProcessStepTaskHandler) HandleCurrentTask() error {
	logrus.Debugf("[%s]handling step[%s]", this.CurentStepTask.GetJobID(), this.StepTaskType)

	this.StartStep()

	if this.CurentStepTask.IsJobRollbacking() {
		this.Rollback()
	} else {
		// Simulate the processing of current order step
		if !this.CurentStepTask.IsJobInFinishingStep() {
			time.Sleep(time.Second * StepProcessTime)
		}

		// Simulate the processing failure occurs at 5% ratio
		if util.IsEventWithSpecifiedRatioHappens() && !this.CurentStepTask.IsJobInFinishingStep() {
			this.CurentStepTask.MarkJobAsFailure()

			logrus.Debugf("[%s]Failure occurs when handling step[%s]",
				this.CurentStepTask.GetJobID(), this.CurentStepTask.GetCurrentStep())
		} else {
			this.FinishStep()
		}
	}

	this.PipeLine.DispatchTask(this.CurentStepTask.GetJobID())
	return nil
}

func (this *ProcessStepTaskHandler) AppendTask(job IJob) {
	this.PendingTasks <- job.(*ProcessJob)
}

func (this *ProcessStepTaskHandler) Rollback() {
	logrus.Debugf("[%s]Rollback step[%s]", this.CurentStepTask.GetJobID(), this.StepTaskType)

	this.CurentStepTask.RollbackStep(this.StepTaskType)

	this.CurentStepTask.UpdateDatabase()
}

func (this *ProcessStepTaskHandler) StartStep() error {
	job := this.CurentStepTask
	err := VerifyStepSwitch(job.GetCurrentStep(), this.StepTaskType)
	if err != nil {
		return err
	}

	job.StartStep(this.StepTaskType)

	job.UpdateDatabase()

	logrus.Debugf("[%s]Start step[%s]", job.GetJobID(), job.GetCurrentStep())
	return nil
}

func (this *ProcessStepTaskHandler) FinishStep() error {
	job := this.CurentStepTask
	if this.StepTaskType != job.GetCurrentStep() {
		return errors.New("Cannot finish step since it is not current step")
	}

	job.FinishCurrentStep()

	if job.IsFailureOccured() && !job.IsJobRollbacking() {
		job.StartRollback()
	}

	job.UpdateDatabase()

	logrus.Debugf("[%s]Finish step[%s]", job.GetJobID(), job.GetCurrentStep())
	return nil
}
