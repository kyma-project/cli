package kubeconfig

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_parseExpirationTime(t *testing.T) {
	tests := []struct {
		name            string
		time            string
		expectedSeconds int64
		expectedError   bool
	}{
		{
			name:            "1 hour",
			time:            "1h",
			expectedSeconds: 3600,
		},
		{
			name:            "1 day",
			time:            "1d",
			expectedSeconds: 86400,
		},
		{
			name:            "1 day 1 hour",
			time:            "1d1h",
			expectedSeconds: 0,
			expectedError:   true,
		},
		{
			name:            "empty string",
			time:            "",
			expectedSeconds: 0,
			expectedError:   true,
		},
	}

	for _, tt := range tests {
		time := tt.time
		expectedError := tt.expectedError
		expectedSeconds := tt.expectedSeconds
		t.Run(tt.name, func(t *testing.T) {
			seconds, err := parseExpirationTime(time)
			if expectedError {
				require.NotNil(t, err)
				return
			} else {
				require.Nil(t, err)
			}
			require.Equal(t, expectedSeconds, seconds)
		})
	}
}
