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
		defer asyncUI.Stop()
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
		assert.Len(t, mockStepFactory.Steps, 1)
		assert.Contains(t, mockStepFactory.Steps[0].Statuses(), "Installing Kyma")
	})

	t.Run("Send start and stop event with success", func(t *testing.T) {
		mockStepFactory := &StepFactoryMock{}
		asyncUI := AsyncUI{StepFactory: mockStepFactory}
		updCh := asyncUI.Start()
		defer asyncUI.Stop()
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
		assert.Len(t, mockStepFactory.Steps, 1)
		assert.True(t, mockStepFactory.Steps[0].IsSuccessful())
	})

	t.Run("Send start and stop events with failure", func(t *testing.T) {
		mockStepFactory := &StepFactoryMock{}
		asyncUI := AsyncUI{StepFactory: mockStepFactory}
		updCh := asyncUI.Start()
		defer asyncUI.Stop()
		updCh <- deployment.ProcessUpdate{
			Event:     deployment.ProcessStart,
			Phase:     deployment.InstallPreRequisites,
			Component: components.KymaComponent{},
		}
		updCh <- deployment.ProcessUpdate{
			Event:     deployment.ProcessFinished,
			Phase:     deployment.InstallPreRequisites,
			Component: components.KymaComponent{},
		}
		updCh <- deployment.ProcessUpdate{
			Event:     deployment.ProcessStart,
			Phase:     deployment.InstallComponents,
			Component: components.KymaComponent{},
		}
		updCh <- deployment.ProcessUpdate{
			Event:     deployment.ProcessForceQuitFailure,
			Phase:     deployment.InstallComponents,
			Component: components.KymaComponent{},
		}
		assert.Len(t, mockStepFactory.Steps, 2)
		assert.True(t, mockStepFactory.Steps[0].IsSuccessful())
		assert.False(t, mockStepFactory.Steps[1].IsSuccessful())
	})
}
