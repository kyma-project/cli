package module

import (
	"github.com/kyma-project/lifecycle-manager/api/v1beta2"
	"testing"
)

func TestValidateVersion(t *testing.T) {
	tests := []struct {
		tCase   string
		version string
		isValid bool
	}{
		{
			tCase:   "proper version without 'v' prefix",
			version: "1.2.3",
			isValid: true,
		},
		{
			tCase:   "proper version with 'v' prefix",
			version: "v1.2.3",
			isValid: true,
		},
		{
			tCase:   "invalid version - missing patch",
			version: "v1.2",
			isValid: false,
		},
		{
			tCase:   "invalid version - just a string",
			version: "main",
			isValid: false,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.tCase, func(t *testing.T) {
				mc := Config{
					Version: tt.version,
				}

				err := newConfigValidator(&mc).
					validateVersion().
					do()

				if !tt.isValid && err == nil {
					t.Errorf("Version validation failed for input %q: Expected an error but success is reported", tt.version)
				}
				if tt.isValid && err != nil {
					t.Error(err)
				}
			},
		)
	}
}

func TestValidateName(t *testing.T) {
	tests := []struct {
		tCase   string
		name    string
		isValid bool
	}{
		{
			tCase:   "proper name",
			name:    "kyma-project.io/foo",
			isValid: true,
		},
		{
			tCase:   "empty name",
			name:    "",
			isValid: false,
		},
		{
			tCase:   "invalid name - whitespaces",
			name:    " kyma-project.io/foo ",
			isValid: false,
		},
		{
			tCase:   "invalid name - no path",
			name:    "kyma-project.io",
			isValid: false,
		},
		{
			tCase:   "invalid name",
			name:    "foo",
			isValid: false,
		},
		{
			tCase:   "invalid name with path",
			name:    "foo/bar",
			isValid: false,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.tCase, func(t *testing.T) {
				mc := Config{
					Name: tt.name,
				}

				err := newConfigValidator(&mc).
					validateName().
					do()

				if !tt.isValid && err == nil {
					t.Errorf("Name validation failed for input %q: Expected an error but success is reported", tt.name)
				}
				if tt.isValid && err != nil {
					t.Error(err)
				}
			},
		)
	}
}

func TestValidateNamespace(t *testing.T) {
	tests := []struct {
		tCase     string
		namespace string
		isValid   bool
	}{
		{
			tCase:     "proper namespace",
			namespace: "kyma-system",
			isValid:   true,
		},
		{
			tCase:     "proper namespace",
			namespace: "kyma-system-1",
			isValid:   true,
		},
		{
			tCase:     "empty namespace",
			namespace: "",
			isValid:   true,
		},
		{
			tCase:     "invalid namespace - contains capital letters",
			namespace: "Kyma",
			isValid:   false,
		},
		{
			tCase:     "invalid namespace - contains invalid characters",
			namespace: "kyma,system",
			isValid:   false,
		},
		{
			tCase:     "invalid namespace - starts with hyphen",
			namespace: "-kyma",
			isValid:   false,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.tCase, func(t *testing.T) {
				mc := Config{
					Namespace: tt.namespace,
				}

				err := newConfigValidator(&mc).
					validateNamespace().
					do()

				if !tt.isValid && err == nil {
					t.Errorf("Namespace validation failed for input %q: Expected an error but success is reported", tt.namespace)
				}
				if tt.isValid && err != nil {
					t.Error(err)
				}
			},
		)
	}
}

func TestValidateChannel(t *testing.T) {
	tests := []struct {
		tCase   string
		channel string
		isValid bool
	}{
		{
			tCase:   "proper channel",
			channel: "regular",
			isValid: true,
		},
		{
			tCase:   "empty channel",
			channel: "",
			isValid: false,
		},
		{
			tCase:   "channel too short",
			channel: "ab",
			isValid: false,
		},
		{
			tCase:   "channel too long",
			channel: "thisstringvaluehaslengthof33chars",
			isValid: false,
		},
		{
			tCase:   "channel contains invalid characters",
			channel: "this value has spaces",
			isValid: false,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.tCase, func(t *testing.T) {
				mc := Config{
					Channel: tt.channel,
				}

				err := newConfigValidator(&mc).
					validateChannel().
					do()

				if !tt.isValid && err == nil {
					t.Errorf("Channel validation failed for input %q: Expected an error but success is reported", tt.channel)
				}
				if tt.isValid && err != nil {
					t.Error(err)
				}
			},
		)
	}
}

func TestValidateCustomStateChecks(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "No error when custom state check not provided",
			config: Config{
				CustomStateChecks: nil,
			},
			wantErr: false,
		},
		{
			name: "Error when only one valid custom state check provided",
			config: Config{
				CustomStateChecks: []v1beta2.CustomStateCheck{
					{
						JSONPath:    "status.health",
						Value:       "green",
						MappedState: v1beta2.StateReady,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Error when custom state check without JSONPath provided",
			config: Config{
				CustomStateChecks: []v1beta2.CustomStateCheck{
					{
						JSONPath:    "status.health",
						Value:       "red",
						MappedState: v1beta2.StateError,
					},
					{
						JSONPath:    "",
						Value:       "green",
						MappedState: v1beta2.StateReady,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Error when custom state check without MappedState provided",
			config: Config{
				CustomStateChecks: []v1beta2.CustomStateCheck{
					{
						JSONPath:    "status.health",
						Value:       "green",
						MappedState: v1beta2.StateReady,
					},
					{
						JSONPath:    "status.health",
						Value:       "red",
						MappedState: "",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "No error when at least two valid custom state checks provided",
			config: Config{
				CustomStateChecks: []v1beta2.CustomStateCheck{
					{
						JSONPath:    "status.health",
						Value:       "green",
						MappedState: v1beta2.StateReady,
					},
					{
						JSONPath:    "status.health",
						Value:       "red",
						MappedState: v1beta2.StateError,
					},
					{
						JSONPath:    "status.health",
						Value:       "yellow",
						MappedState: v1beta2.StateWarning,
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cv := &configValidator{
				config:     &tt.config,
				validators: []configValidationFunc{},
			}
			err := cv.validateCustomStateChecks().do()
			if err == nil && tt.wantErr {
				t.Error("validateCustomStateChecks() returned no error when an error was expected")
			}
			if err != nil && !tt.wantErr {
				t.Errorf("validateCustomStateChecks() returned %v, when no error was expected", err)
			}
		})
	}
}
