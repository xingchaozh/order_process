package pipeline

import (
	"errors"

	"github.com/Sirupsen/logrus"
)

// The inteface of pipeline for Order Processing Service
type IPipeline interface {
	// Start the pipeline
	Start()
	// Append new job to pipeline
	AppendJob(job IJob)
	// Dispatch task to next step
	DispatchTask(jobId string)
}

var ProcessSteps = []string{
	"Scheduling",
	"Pre-Processing",
	"Processing",
	"Post-Processing",
	"Completed",
	"Failed",
}

var StepsSwitchMap = map[string][]string{
	"Scheduling":      {"Scheduling", "Pre-Processing", "Failed"},
	"Pre-Processing":  {"Pre-Processing", "Processing", "Failed"},
	"Processing":      {"Processing", "Post-Processing", "Failed"},
	"Post-Processing": {"Post-Processing", "Completed", "Failed"},
	"Completed":       {"Completed"},
	"Failed":          {"Failed"},
}

var StepsSwitchNextMap = map[string][]string{
	"Scheduling":      {"Pre-Processing", "Failed"},
	"Pre-Processing":  {"Processing", "Failed"},
	"Processing":      {"Post-Processing", "Failed"},
	"Post-Processing": {"Completed", "Failed"},
	"Completed":       {"Completed"},
	"Failed":          {"Failed"},
}

const (
	MaxProcessJobsCountPerPipeline = 10000
)

// The definition of Order Processing Pipeline
type ProcessPipeline struct {
	Jobs         map[string]IJob
	TaskHandlers map[string]ITaskHandler
}

// The constructor of pipeline for Order Processing Service
func NewProcessPipeline(NewTaskHandler func(string, IPipeline) ITaskHandler) IPipeline {
	pipeline := ProcessPipeline{
		Jobs:         make(map[string]IJob),
		TaskHandlers: make(map[string]ITaskHandler),
	}
	for _, stepName := range ProcessSteps {
		pipeline.TaskHandlers[stepName] = NewTaskHandler(stepName, &pipeline)
	}
	return &pipeline
}

// Start the pipeline
func (this *ProcessPipeline) Start() {
	for _, handler := range this.TaskHandlers {
		go handler.PerformTasks()
	}
}

// Append process job to pipeline
func (this *ProcessPipeline) AppendJob(job IJob) {
	if _, found := this.Jobs[job.GetJobID()]; found {
		logrus.Errorf("ProcessJob existed:[%v]", job.GetJobID())
		return
	}
	// Insert job
	this.Jobs[job.GetJobID()] = job

	// Schedule the job immediately
	this.DispatchTask(job.GetJobID())
}

// Dispatch the task
func (this *ProcessPipeline) DispatchTask(jobId string) {
	job := this.Jobs[jobId]
	if job.IsJobInFinishingStep() && !job.IsJobRollbacking() {
		this.FinishOrder(jobId)
		return
	}

	nextStep, err := this.GetNextStep(jobId)
	if err != nil {
		logrus.Errorf("[%s]DispatchStepTask,current step: [%s], error:[%v]",
			job.GetJobID(), job.GetCurrentStep(), err)
		return
	}

	logrus.Debugf("[%s]DispatchStepTask,current step: [%s], next step:[%s]",
		job.GetJobID(), job.GetCurrentStep(), nextStep)

	this.TaskHandlers[nextStep].AppendTask(job)
}

// Get next processing step
func (this *ProcessPipeline) GetNextStep(jobId string) (string, error) {
	job := this.Jobs[jobId]

	if !job.IsCurrentStepCompleted() && !job.IsFailureOccured() {
		return job.GetCurrentStep(), nil
	}

	if job.IsJobRollbacking() && job.IsJobInFinishingStep() {
		nextRollbackStep, err := job.GetRollbackStep()
		if err != nil {
			return "", err
		}
		return nextRollbackStep, nil
	}

	if nextSteps, found := StepsSwitchNextMap[job.GetCurrentStep()]; found {
		if !job.IsJobInFinishingStep() {
			if !job.IsFailureOccured() {
				return nextSteps[0], nil
			} else {
				return nextSteps[1], nil
			}
		}
		return nextSteps[0], nil
	}

	return "", errors.New("cannot find next step")
}

// Finish the order
func (this *ProcessPipeline) FinishOrder(jobId string) {
	logrus.Debugf("[%s]Finish Order", jobId)
	logrus.Debugln()
	this.Jobs[jobId].FinalizeJob()
}

// Verify whether the step switch is valid
func VerifyStepSwitch(previousStep string, currentStep string) error {
	for _, step := range StepsSwitchMap[previousStep] {
		if step == currentStep {
			return nil
		}
	}
	return errors.New("Step switch verification failed")
}
