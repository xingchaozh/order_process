package pipeline

import (
	"errors"
	"sync"

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
	// Stop the pipeline
	Stop()
}

type ProcessStep int

const (
	Scheduling ProcessStep = iota
	PreProcessing
	Processing
	PostProcessing
	Completed
	Failed
	Unknown
)

var ProcessSteps = []ProcessStep{
	Scheduling,
	PreProcessing,
	Processing,
	PostProcessing,
	Completed,
	Failed,
}

var ProcessStepNames = map[ProcessStep]string{
	Scheduling:     "Scheduling",
	PreProcessing:  "Pre-Processing",
	Processing:     "Processing",
	PostProcessing: "Post-Processing",
	Completed:      "Completed",
	Failed:         "Failed",
}

func (s ProcessStep) String() string {
	return ProcessStepNames[s]
}

func ParseProcessStep(step string) ProcessStep {
	if step == Scheduling.String() {
		return Scheduling
	} else if step == PreProcessing.String() {
		return PreProcessing
	} else if step == Processing.String() {
		return Processing
	} else if step == PostProcessing.String() {
		return PostProcessing
	} else if step == Completed.String() {
		return Completed
	} else if step == Failed.String() {
		return Failed
	} else {
		logrus.Error("Parse ProcessStep failed")
		return Unknown
	}
}

var StepsSwitchMap = map[ProcessStep][]ProcessStep{
	Scheduling:     {Scheduling, PreProcessing, Failed},
	PreProcessing:  {PreProcessing, Processing, Failed},
	Processing:     {Processing, PostProcessing, Failed},
	PostProcessing: {PostProcessing, Completed, Failed},
	Completed:      {Completed},
	Failed:         {Failed},
}

var StepsSwitchNextMap = map[ProcessStep][]ProcessStep{
	Scheduling:     {PreProcessing, Failed},
	PreProcessing:  {Processing, Failed},
	Processing:     {PostProcessing, Failed},
	PostProcessing: {Completed, Failed},
	Completed:      {Completed},
	Failed:         {Failed},
}

const (
	MaxProcessJobsCountPerPipeline = 10000
)

// The definition of Order Processing Pipeline
type ProcessPipeline struct {
	Jobs         map[string]IJob
	TaskHandlers map[string]ITaskHandler
	lock         sync.Mutex
}

// The constructor of pipeline for Order Processing Service
func NewProcessPipeline(NewTaskHandler func(string, IPipeline) ITaskHandler) IPipeline {
	pipeline := ProcessPipeline{
		Jobs:         make(map[string]IJob),
		TaskHandlers: make(map[string]ITaskHandler),
	}
	for _, step := range ProcessSteps {
		pipeline.TaskHandlers[step.String()] = NewTaskHandler(step.String(), &pipeline)
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
	{
		defer this.lock.Unlock()
		this.lock.Lock()
		// Insert job
		this.Jobs[job.GetJobID()] = job
	}
	// Schedule the job immediately
	this.DispatchTask(job.GetJobID())
}

// Dispatch the task to next task handler
func (this *ProcessPipeline) DispatchTask(jobId string) {
	job := this.Jobs[jobId]
	if job.IsJobInFinishingStep() && !job.IsJobRollbacking() {
		this.FinishJob(jobId)
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

	if !job.IsCurrentStepCompleted() && !job.IsErrorOccured() {
		return job.GetCurrentStep(), nil
	}

	if job.IsJobRollbacking() && job.IsJobInFinishingStep() {
		nextRollbackStep, err := job.GetRollbackStep()
		if err != nil {
			return "", err
		}
		return nextRollbackStep, nil
	}

	if nextSteps, found := StepsSwitchNextMap[ParseProcessStep(job.GetCurrentStep())]; found {
		if !job.IsJobInFinishingStep() {
			if !job.IsErrorOccured() {
				return nextSteps[0].String(), nil
			} else {
				return nextSteps[1].String(), nil
			}
		}
		return nextSteps[0].String(), nil
	}

	return "", errors.New("cannot find next step")
}

// Finalize the order if no more process is needed.
func (this *ProcessPipeline) FinishJob(jobId string) {
	logrus.Debugf("[%s]Finish Order", jobId)
	logrus.Debugln()
	this.Jobs[jobId].FinalizeJob()

	{
		defer this.lock.Unlock()
		this.lock.Lock()
		// Remove job from cached mapping
		delete(this.Jobs, jobId)
	}
}

// Stop the pipeline
func (this *ProcessPipeline) Stop() {
	for _, handler := range this.TaskHandlers {
		handler.Stop()
	}
}

// Verify whether the step switch is valid
func VerifyStepSwitch(previousStep string, currentStep string) error {
	for _, step := range StepsSwitchMap[ParseProcessStep(previousStep)] {
		if step.String() == currentStep {
			return nil
		}
	}
	return errors.New("Step switch verification failed")
}
