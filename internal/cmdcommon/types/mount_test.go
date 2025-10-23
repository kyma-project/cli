package types

import (
	"testing"
)

func TestMountArray_Set_NormalFormat(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    MountSpec
		expectError bool
	}{
		{
			name:  "backward compatibility - just name",
			input: "my-secret",
			expected: MountSpec{
				Name:     "my-secret",
				Path:     "",
				Key:      "",
				ReadOnly: false,
			},
			expectError: false,
		},
		{
			name:  "full normal format",
			input: "name=my-secret,path=/app/config,key=config.yaml,ro=true",
			expected: MountSpec{
				Name:     "my-secret",
				Path:     "/app/config",
				Key:      "config.yaml",
				ReadOnly: true,
			},
			expectError: false,
		},
		{
			name:  "normal format without key",
			input: "name=my-configmap,path=/app/data,ro=false",
			expected: MountSpec{
				Name:     "my-configmap",
				Path:     "/app/data",
				Key:      "",
				ReadOnly: false,
			},
			expectError: false,
		},
		{
			name:  "normal format without readonly",
			input: "name=tls-secret,path=/etc/certs,key=tls.crt",
			expected: MountSpec{
				Name:     "tls-secret",
				Path:     "/etc/certs",
				Key:      "tls.crt",
				ReadOnly: false,
			},
			expectError: false,
		},
		{
			name:        "missing name",
			input:       "path=/app/config,key=config.yaml",
			expectError: true,
		},
		{
			name:        "missing path when other fields specified",
			input:       "name=my-secret,key=config.yaml",
			expectError: true,
		},
		{
			name:        "invalid field",
			input:       "name=my-secret,path=/app,invalid=value",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MountArray{}
			err := m.Set(tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if len(m.Mounts) != 1 {
				t.Errorf("expected 1 mount, got %d", len(m.Mounts))
				return
			}

			got := m.Mounts[0]
			if got.Name != tt.expected.Name {
				t.Errorf("expected name %q, got %q", tt.expected.Name, got.Name)
			}
			if got.Path != tt.expected.Path {
				t.Errorf("expected path %q, got %q", tt.expected.Path, got.Path)
			}
			if got.Key != tt.expected.Key {
				t.Errorf("expected key %q, got %q", tt.expected.Key, got.Key)
			}
			if got.ReadOnly != tt.expected.ReadOnly {
				t.Errorf("expected readonly %v, got %v", tt.expected.ReadOnly, got.ReadOnly)
			}
		})
	}
}

func TestMountArray_Set_ShorthandFormat(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    MountSpec
		expectError bool
	}{
		{
			name:  "shorthand with key and readonly",
			input: "tls-secret:tls.crt=/etc/certs:ro",
			expected: MountSpec{
				Name:     "tls-secret",
				Path:     "/etc/certs",
				Key:      "tls.crt",
				ReadOnly: true,
			},
			expectError: false,
		},
		{
			name:  "shorthand without key and readonly",
			input: "app-config=/app/config",
			expected: MountSpec{
				Name:     "app-config",
				Path:     "/app/config",
				Key:      "",
				ReadOnly: false,
			},
			expectError: false,
		},
		{
			name:  "shorthand with key but not readonly",
			input: "my-secret:password=/app/secrets",
			expected: MountSpec{
				Name:     "my-secret",
				Path:     "/app/secrets",
				Key:      "password",
				ReadOnly: false,
			},
			expectError: false,
		},
		{
			name:  "shorthand without key but with readonly",
			input: "my-configmap=/app/data:ro",
			expected: MountSpec{
				Name:     "my-configmap",
				Path:     "/app/data",
				Key:      "",
				ReadOnly: true,
			},
			expectError: false,
		},
		{
			name:        "shorthand missing path",
			input:       "my-secret:key=",
			expectError: true,
		},
		{
			name:        "shorthand missing path",
			input:       "my-secret:key",
			expectError: true,
		},
		{
			name:        "shorthand empty name",
			input:       ":key=/path",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MountArray{}
			err := m.Set(tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if len(m.Mounts) != 1 {
				t.Errorf("expected 1 mount, got %d", len(m.Mounts))
				return
			}

			got := m.Mounts[0]
			if got.Name != tt.expected.Name {
				t.Errorf("expected name %q, got %q", tt.expected.Name, got.Name)
			}
			if got.Path != tt.expected.Path {
				t.Errorf("expected path %q, got %q", tt.expected.Path, got.Path)
			}
			if got.Key != tt.expected.Key {
				t.Errorf("expected key %q, got %q", tt.expected.Key, got.Key)
			}
			if got.ReadOnly != tt.expected.ReadOnly {
				t.Errorf("expected readonly %v, got %v", tt.expected.ReadOnly, got.ReadOnly)
			}
		})
	}
}

func TestMountArray_Set_PathTraversal(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
	}{
		{
			name:        "path traversal with ..",
			input:       "name=secret,path=../etc/passwd",
			expectError: true,
		},
		{
			name:        "path traversal with ../../",
			input:       "name=secret,path=../../root",
			expectError: true,
		},
		{
			name:        "path traversal in shorthand",
			input:       "secret=/../../etc/passwd",
			expectError: true,
		},
		{
			name:        "valid path",
			input:       "name=secret,path=/app/config",
			expectError: false,
		},
		{
			name:        "valid relative path",
			input:       "name=secret,path=config",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MountArray{}
			err := m.Set(tt.input)

			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestMountArray_Set_MixedFormats(t *testing.T) {
	tests := []struct {
		name        string
		inputs      []string
		expectError bool
	}{
		{
			name:        "mixing normal and shorthand",
			inputs:      []string{"name=secret,path=/app", "secret2=/path"},
			expectError: false, // mixing is now allowed since we don't track format
		},
		{
			name:        "mixing shorthand and normal",
			inputs:      []string{"secret=/path", "name=secret2,path=/app"},
			expectError: false, // mixing is now allowed since we don't track format
		},
		{
			name:        "all normal format",
			inputs:      []string{"name=secret1,path=/app1", "name=secret2,path=/app2"},
			expectError: false,
		},
		{
			name:        "all shorthand format",
			inputs:      []string{"secret1=/app1", "secret2=/app2"},
			expectError: false,
		},
		{
			name:        "backward compatibility mixed with normal",
			inputs:      []string{"old-secret", "name=new-secret,path=/app"},
			expectError: false, // backward compatibility should work with normal format
		},
		{
			name:        "invalid shorthand format in sequence",
			inputs:      []string{"secret1=/app1", "invalid:format"},
			expectError: true, // second input has colon but no equals - should error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MountArray{}
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestMountArray_String(t *testing.T) {
	tests := []struct {
		name     string
		mounts   []MountSpec
		expected string
	}{
		{
			name:     "empty mounts",
			mounts:   []MountSpec{},
			expected: "",
		},
		{
			name: "single mount with all fields",
			mounts: []MountSpec{
				{Name: "secret", Path: "/app", Key: "key", ReadOnly: true},
			},
			expected: "name=secret,path=/app,key=key,ro=true",
		},
		{
			name: "single mount without key",
			mounts: []MountSpec{
				{Name: "secret", Path: "/app", ReadOnly: false},
			},
			expected: "name=secret,path=/app,ro=false",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MountArray{Mounts: tt.mounts}
			got := m.String()
			if got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestServiceBindingSecretArray_Set(t *testing.T) {
	tests := []struct {
		name     string
		inputs   []string
		expected []string
	}{
		{
			name:     "single secret",
			inputs:   []string{"secret1"},
			expected: []string{"secret1"},
		},
		{
			name:     "multiple secrets",
			inputs:   []string{"secret1", "secret2", "secret3"},
			expected: []string{"secret1", "secret2", "secret3"},
		},
		{
			name:     "empty input",
			inputs:   []string{""},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ServiceBindingSecretArray{}
			for _, input := range tt.inputs {
				err := s.Set(input)
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}

			if len(s.Names) != len(tt.expected) {
				t.Errorf("expected %d names, got %d", len(tt.expected), len(s.Names))
				return
			}

			for i, expected := range tt.expected {
				if s.Names[i] != expected {
					t.Errorf("expected name[%d] %q, got %q", i, expected, s.Names[i])
				}
			}
		})
	}
}
