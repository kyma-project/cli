package asyncui

import (
	"context"
	"fmt"

	"github.com/kyma-incubator/hydroform/parallel-install/pkg/components"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/deployment"
	"github.com/kyma-project/cli/pkg/step"
)

//End-user messages
const (
	deployPrerequisitesPhaseMsg   string = "Deploying pre-requisites"
	undeployPrerequisitesPhaseMsg string = "Undeploying pre-requisites"
	deployComponentsPhaseMsg      string = "Deploying Kyma"
	undeployComponentsPhaseMsg    string = "Undeploying Kyma"
	deployComponentMsg            string = "Component '%s' deployed"
	undeployComponentMsg          string = "Component '%s' removed"
)

//AsyncUI renders the CLI ui based on receiving events
type AsyncUI struct {
	//used to create UI steps
	StepFactory step.FactoryInterface
	//processing context
	context context.Context
	//cancel function the caller can execute to interrupt processing
	Cancel context.CancelFunc
	//channel to retrieve update events
	updates chan deployment.ProcessUpdate
	//channel to pass errors to caller
	Errors chan error
	//internal state
	running bool
	//a failure occurred
	Failed bool
}

//Start renders the CLI UI and provides the channel for receiving events
func (ui *AsyncUI) Start() error {
	if ui.running {
		return fmt.Errorf("Duplicate call of start method detected")
	}
	ui.running = true

	//process async process updates
	ui.updates = make(chan deployment.ProcessUpdate)
	//initialize processing context
	ui.context, ui.Cancel = context.WithCancel(context.Background())

	go func() {
		defer ui.Cancel()
		ongoingSteps := make(map[deployment.InstallationPhase]step.Step)
		for procUpdateEvent := range ui.updates {
			switch procUpdateEvent.Event {
			case deployment.ProcessRunning:
				//dispatch only component related ProcessRunning events
				//(they provide information about the component installation result)
				if procUpdateEvent.IsComponentUpdate() {
					ui.dispatchError(ui.renderStopEvent(procUpdateEvent, &ongoingSteps))
				}
				continue
			case deployment.ProcessStart:
				ui.dispatchError(ui.renderStartEvent(procUpdateEvent, &ongoingSteps))
			default:
				ui.dispatchError(ui.renderStopEvent(procUpdateEvent, &ongoingSteps))
			}
		}
	}()

	return nil
}

//dispatchError will pass an error to the Caller
func (ui *AsyncUI) dispatchError(err error) {
	if err != nil {
		ui.Failed = true
		//fire error event to caller's error channel
		if ui.Errors != nil {
			ui.Errors <- err
		}
	}
}

//Stop will close the update channel and wait until the the UI rendering is finished
func (ui *AsyncUI) Stop() {
	if !ui.running {
		return
	}
	close(ui.updates)
	<-ui.context.Done()
	ui.running = false
}

//renderStartEvent dispatches a start event to an UI step
func (ui *AsyncUI) renderStartEvent(procUpdEvent deployment.ProcessUpdate, ongoingSteps *map[deployment.InstallationPhase]step.Step) error {
	if _, exists := (*ongoingSteps)[procUpdEvent.Phase]; exists {
		return fmt.Errorf("Illegal state: start-step for installation phase '%s' already exists", procUpdEvent.Phase)
	}
	step := ui.StepFactory.NewStep(ui.majorStepMsg(procUpdEvent))
	step.Start()
	(*ongoingSteps)[procUpdEvent.Phase] = step
	return nil
}

func (ui *AsyncUI) majorStepMsg(procUpdEvent deployment.ProcessUpdate) string {
	//create a major step
	var stepMsg string
	switch procUpdEvent.Phase {
	case deployment.InstallPreRequisites:
		stepMsg = deployPrerequisitesPhaseMsg
	case deployment.UninstallPreRequisites:
		stepMsg = undeployPrerequisitesPhaseMsg
	case deployment.InstallComponents:
		stepMsg = deployComponentsPhaseMsg
	case deployment.UninstallComponents:
		stepMsg = undeployComponentsPhaseMsg
	default:
		//non-deployment specific installation phase
		//e.g. steps triggered by CLI before or after the deployment
		stepMsg = string(procUpdEvent.Phase)
	}
	return stepMsg
}

//renderStopEvent dispatches a stop event
func (ui *AsyncUI) renderStopEvent(procUpdEvent deployment.ProcessUpdate, ongoingSteps *map[deployment.InstallationPhase]step.Step) error {
	var err error
	if procUpdEvent.IsComponentUpdate() {
		//event is related to a component
		err = ui.renderStopEventComponent(procUpdEvent)
	} else {
		//event is related to a major installation phases
		err = ui.renderStopEventInstallationPhase(procUpdEvent, ongoingSteps)
	}
	return err
}

//renderStopEventInstallationPhase stops the existing step of the installation phase
func (ui *AsyncUI) renderStopEventInstallationPhase(procUpdEvent deployment.ProcessUpdate, ongoingSteps *map[deployment.InstallationPhase]step.Step) error {
	//for convenience
	event := procUpdEvent.Event
	installPhase := procUpdEvent.Phase
	err := procUpdEvent.Error

	if _, exists := (*ongoingSteps)[installPhase]; !exists {
		return fmt.Errorf("Illegal state: step for installation phase '%s' does not exist", installPhase)
	}

	//all good
	if event == deployment.ProcessFinished {
		(*ongoingSteps)[installPhase].Successf("%s finished successfully", ui.majorStepMsg(procUpdEvent))
		return nil
	}

	//something went wrong
	errMsg := fmt.Sprintf("%s failed", ui.majorStepMsg(procUpdEvent))
	if err != nil {
		errMsg = fmt.Sprintf("%s\n%s", errMsg, err)
	}
	(*ongoingSteps)[installPhase].Failuref(errMsg)

	return fmt.Errorf("Deployment phase '%s' failed: %s\n%v", installPhase, event, err)
}

//renderStopEventComponent displays a component stop event in a new step
func (ui *AsyncUI) renderStopEventComponent(procUpdEvent deployment.ProcessUpdate) error {
	//for convenience
	comp := procUpdEvent.Component
	installPhase := procUpdEvent.Phase
	cmpErr := comp.Error

	//determine step name
	var stepName string
	if installPhase == deployment.InstallComponents {
		stepName = fmt.Sprintf(deployComponentMsg, comp.Name)
	} else {
		stepName = fmt.Sprintf(undeployComponentMsg, comp.Name)
	}

	//create step for processed component
	step := ui.StepFactory.NewStep(stepName)
	if comp.Status == components.StatusError {
		errMsg := fmt.Sprintf("Deployment of component '%s' failed", comp.Name)
		if cmpErr != nil {
			errMsg = fmt.Sprintf("%s\n%s", errMsg, cmpErr)
		}
		step.Failuref(errMsg)
		return fmt.Errorf(errMsg)
	}
	step.Success()

	return nil
}

//AddStep adds an additional installation step
func (ui *AsyncUI) AddStep(step string) (step.Step, error) {
	if !ui.running {
		return nil, fmt.Errorf("Cannot add an step because AsyncUI is not running")
	}
	return ui.StepFactory.NewStep(step), nil
}

//UpdateChannel returns the update channel which retrieves process update events
func (ui *AsyncUI) UpdateChannel() (chan<- deployment.ProcessUpdate, error) {
	if !ui.running {
		return nil, fmt.Errorf("Update channel cannot be retrieved because AsyncUI is not running")
	}
	return ui.updates, nil
}

//IsRunning returns true if the AsyncUI is still receiving events
func (ui *AsyncUI) IsRunning() bool {
	return ui.running
}
