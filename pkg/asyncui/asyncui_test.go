package asyncui

import (
	"testing"

	"github.com/kyma-incubator/hydroform/parallel-install/pkg/components"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/deployment"
	"github.com/kyma-project/cli/pkg/step"
	stepMocks "github.com/kyma-project/cli/pkg/step/mocks"
	"github.com/stretchr/testify/assert"
)

type StepFactoryMock struct {
	Steps []*stepMocks.Step
}

func (mock *StepFactoryMock) NewStep(msg string) step.Step {
	step := &stepMocks.Step{}
	step.Status(msg)
	mock.Steps = append(mock.Steps, step)
	return step
}

func TestFailedComponent(t *testing.T) {
	t.Parallel()

	t.Run("Send duplicate start events", func(t *testing.T) {
		mockStepFactory := &StepFactoryMock{}
		asyncUI := AsyncUI{StepFactory: mockStepFactory}
		updCh := asyncUI.Start()
		updCh <- deployment.ProcessUpdate{
			Event:     deployment.ProcessStart,
			Phase:     deployment.InstallComponents,
			Component: components.KymaComponent{},
		}
		// duplicate start events have to be ignored
		updCh <- deployment.ProcessUpdate{
			Event:     deployment.ProcessStart,
			Phase:     deployment.InstallComponents,
			Component: components.KymaComponent{},
		}
		asyncUI.Stop() //stop receiving events and wait until processing is finished
		assert.Len(t, mockStepFactory.Steps, 1)
		assert.Contains(t, mockStepFactory.Steps[0].Statuses(), "Installing Kyma")
	})

	t.Run("Send start and stop event with success", func(t *testing.T) {
		mockStepFactory := &StepFactoryMock{}
		asyncUI := AsyncUI{StepFactory: mockStepFactory}
		updCh := asyncUI.Start()
		updCh <- deployment.ProcessUpdate{
			Event:     deployment.ProcessStart,
			Phase:     deployment.InstallComponents,
			Component: components.KymaComponent{},
		}
		updCh <- deployment.ProcessUpdate{
			Event:     deployment.ProcessFinished,
			Phase:     deployment.InstallComponents,
			Component: components.KymaComponent{},
		}
		asyncUI.Stop() //stop receiving events and wait until processing is finished
		assert.Len(t, mockStepFactory.Steps, 1)
		assert.True(t, mockStepFactory.Steps[0].IsSuccessful())
	})

	t.Run("Send start and stop events with failure", func(t *testing.T) {
		mockStepFactory := &StepFactoryMock{}
		asyncUI := AsyncUI{StepFactory: mockStepFactory}
		updCh := asyncUI.Start()
		// add step 1 (major installation step)
		updCh <- deployment.ProcessUpdate{
			Event:     deployment.ProcessStart,
			Phase:     deployment.InstallPreRequisites,
			Component: components.KymaComponent{},
		}
		// set status of step 1 to success
		updCh <- deployment.ProcessUpdate{
			Event:     deployment.ProcessFinished,
			Phase:     deployment.InstallPreRequisites,
			Component: components.KymaComponent{},
		}
		// add step 2 (major installation step)
		updCh <- deployment.ProcessUpdate{
			Event:     deployment.ProcessStart,
			Phase:     deployment.InstallComponents,
			Component: components.KymaComponent{},
		}
		// ignore start events related to components
		updCh <- deployment.ProcessUpdate{
			Event: deployment.ProcessStart,
			Phase: deployment.InstallComponents,
			Component: components.KymaComponent{
				Name: "comp1",
			},
		}
		// ignore running events related to components
		updCh <- deployment.ProcessUpdate{
			Event: deployment.ProcessRunning,
			Phase: deployment.InstallComponents,
			Component: components.KymaComponent{
				Name: "comp1",
			},
		}
		// add step 3 with status success (component installation step)
		updCh <- deployment.ProcessUpdate{
			Event: deployment.ProcessFinished,
			Phase: deployment.InstallComponents,
			Component: components.KymaComponent{
				Name:   "comp1",
				Status: components.StatusInstalled,
			},
		}
		// ignore start events related to components
		updCh <- deployment.ProcessUpdate{
			Event: deployment.ProcessStart,
			Phase: deployment.InstallComponents,
			Component: components.KymaComponent{
				Name: "comp2",
			},
		}
		// ignore running events related to components
		updCh <- deployment.ProcessUpdate{
			Event: deployment.ProcessRunning,
			Phase: deployment.InstallComponents,
			Component: components.KymaComponent{
				Name: "comp2",
			},
		}
		// ignore running events related to components
		updCh <- deployment.ProcessUpdate{
			Event: deployment.ProcessRunning,
			Phase: deployment.InstallComponents,
			Component: components.KymaComponent{
				Name: "comp2",
			},
		}
		// add step 4 with status failure (component installation step)
		updCh <- deployment.ProcessUpdate{
			Event: deployment.ProcessExecutionFailure,
			Phase: deployment.InstallComponents,
			Component: components.KymaComponent{
				Name:   "comp2",
				Status: components.StatusError,
			},
		}
		// set status of step 2 to failure
		updCh <- deployment.ProcessUpdate{
			Event:     deployment.ProcessForceQuitFailure,
			Phase:     deployment.InstallComponents,
			Component: components.KymaComponent{},
		}
		asyncUI.Stop() //stop receiving events and wait until processing is finished
		assert.Len(t, mockStepFactory.Steps, 4)
		assert.True(t, mockStepFactory.Steps[0].IsSuccessful())
		assert.False(t, mockStepFactory.Steps[1].IsSuccessful())
		assert.True(t, mockStepFactory.Steps[2].IsSuccessful())
		assert.False(t, mockStepFactory.Steps[3].IsSuccessful())
	})
}
