package module

import (
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
