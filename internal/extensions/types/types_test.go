package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	testActionsMap = map[string]Action{
		"create": nil,
		"debug":  nil,
		"demo":   nil,
	}
)

func TestExtension_Validate(t *testing.T) {
	tests := []struct {
		name      string
		extension Extension
		wantErr   string
	}{
		{
			name: "validation ok",
			extension: Extension{
				Metadata: Metadata{
					Name: "function",
				},
				SubCommands: []Extension{
					{
						Metadata: Metadata{
							Name: "create",
						},
						Action: "create",
						Flags: []Flag{
							{
								Type: "string",
								Name: "test-flag",
							},
						},
						SubCommands: []Extension{
							{
								Metadata: Metadata{
									Name: "demo",
								},
								Action: "demo",
								Args: &Args{
									Type: "bool",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "validation error - broken flag",
			wantErr: "wrong .flags: empty name, unknown type ''\n" +
				"wrong .subCommands[0].metadata: empty name\n" +
				"wrong .subCommands[0].subCommands[1].args: unknown type ''",
			extension: Extension{
				Metadata: Metadata{
					Name: "function",
				},
				// wrong action
				Action: "wrong-action",
				Flags: []Flag{
					{
						// empty flag
					},
				},
				SubCommands: []Extension{
					{
						Metadata: Metadata{
							// no name
						},
						SubCommands: []Extension{
							{
								Metadata: Metadata{
									Name: "demo",
								},
							},
							{
								Metadata: Metadata{
									Name: "create",
								},
								Args: &Args{
									// empty args
								},
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.extension.Validate(testActionsMap)
			if tt.wantErr != "" {
				require.EqualError(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
