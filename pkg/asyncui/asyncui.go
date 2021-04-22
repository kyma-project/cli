package asyncui

import (
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
	//a failure occurred
	Failed bool
}

//Start renders the CLI UI and provides the channel for receiving events
func (ui *AsyncUI) Callback() func(update deployment.ProcessUpdate) {
	ongoingSteps := make(map[deployment.InstallationPhase]step.Step)

	return func(update deployment.ProcessUpdate) {
		switch update.Event {
		case deployment.ProcessRunning:
			//dispatch only component related ProcessRunning events
			//(they provide information about the component installation result)
			if update.IsComponentUpdate() {
				if err := ui.renderStopEvent(update, ongoingSteps); err != nil {
					ui.Failed = true
				}
			}
		case deployment.ProcessStart:
			ui.checkError(ui.renderStartEvent(update, ongoingSteps))
		default:
			ui.checkError(ui.renderStopEvent(update, ongoingSteps))
		}
	}
}

//checkError will pass an error to the Caller
func (ui *AsyncUI) checkError(err error) {
	if err != nil {
		ui.Failed = true
	}
}

//renderStartEvent dispatches a start event to an UI step
func (ui *AsyncUI) renderStartEvent(procUpdEvent deployment.ProcessUpdate, ongoingSteps map[deployment.InstallationPhase]step.Step) error {
	if _, exists := ongoingSteps[procUpdEvent.Phase]; exists {
		return fmt.Errorf("Illegal state: start-step for installation phase '%s' already exists", procUpdEvent.Phase)
	}
	step := ui.StepFactory.NewStep(ui.majorStepMsg(procUpdEvent))
	step.Start()
	ongoingSteps[procUpdEvent.Phase] = step
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
func (ui *AsyncUI) renderStopEvent(procUpdEvent deployment.ProcessUpdate, ongoingSteps map[deployment.InstallationPhase]step.Step) error {
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
func (ui *AsyncUI) renderStopEventInstallationPhase(procUpdEvent deployment.ProcessUpdate, ongoingSteps map[deployment.InstallationPhase]step.Step) error {
	//for convenience
	event := procUpdEvent.Event
	installPhase := procUpdEvent.Phase
	err := procUpdEvent.Error

	if _, exists := ongoingSteps[installPhase]; !exists {
		return fmt.Errorf("Illegal state: step for installation phase '%s' does not exist", installPhase)
	}

	//all good
	if event == deployment.ProcessFinished {
		ongoingSteps[installPhase].Successf("%s finished successfully", ui.majorStepMsg(procUpdEvent))
		return nil
	}

	//something went wrong
	errMsg := fmt.Sprintf("%s failed", ui.majorStepMsg(procUpdEvent))
	if err != nil {
		errMsg = fmt.Sprintf("%s\n%s", errMsg, err)
	}
	ongoingSteps[installPhase].Failuref(errMsg)

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
	if installPhase == deployment.InstallComponents || installPhase == deployment.InstallPreRequisites {
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
	return ui.StepFactory.NewStep(step), nil
}
