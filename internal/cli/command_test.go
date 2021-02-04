package cli

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewStep(t *testing.T) {
	t.Parallel()
	c := Command{} // uninitialized command

	// test uninitialized command
	require.Panics(t, func() { c.NewStep("Oh noes...") }, "NewStep on uninitialized command should panic.")

	c.Options = NewOptions(nil) // properly initialize command

	// test current step update when creating a new step
	s := c.NewStep("test-step")

	require.Equal(t, s, c.CurrentStep, "Command's current step must be the newly created step.")

}
