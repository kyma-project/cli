package types

import (
	"fmt"
	"strings"
)

// MountSpec represents a volume mount specification
type MountSpec struct {
	Name     string
	Path     string
	Key      string
	ReadOnly bool
}

// MountArray holds an array of mount specifications
type MountArray struct {
	Mounts []MountSpec
}

// String returns the string representation
func (m *MountArray) String() string {
	if len(m.Mounts) == 0 {
		return ""
	}
	var parts []string
	for _, mount := range m.Mounts {
		if mount.Key != "" {
			if mount.ReadOnly {
				parts = append(parts, fmt.Sprintf("name=%s,path=%s,key=%s,ro=true", mount.Name, mount.Path, mount.Key))
			} else {
				parts = append(parts, fmt.Sprintf("name=%s,path=%s,key=%s,ro=false", mount.Name, mount.Path, mount.Key))
			}
		} else {
			if mount.ReadOnly {
				parts = append(parts, fmt.Sprintf("name=%s,path=%s,ro=true", mount.Name, mount.Path))
			} else {
				parts = append(parts, fmt.Sprintf("name=%s,path=%s,ro=false", mount.Name, mount.Path))
			}
		}
	}
	return strings.Join(parts, ";")
}

// Type returns the type name
func (m *MountArray) Type() string {
	return "stringArray"
}

// isShorthandFormat determines if a value is in shorthand format
func isShorthandFormat(value string) bool {
	// If it contains comma, it's definitely normal format (multiple key=value pairs)
	if strings.Contains(value, ",") {
		return false
	}
	if !strings.Contains(value, "=") {
		// No equals sign - backward compatibility (just a name)
		return false
	}

	// Has equals but no comma - could be shorthand or simple normal format
	// Check if it follows shorthand pattern: either "name=path" or "name:key=path"
	eqParts := strings.SplitN(value, "=", 2)
	if len(eqParts) == 2 {
		leftSide := eqParts[0]
		rightSide := eqParts[1]

		// If left side contains a colon, it's likely shorthand (name:key=path)
		// If right side looks like a path (starts with / or doesn't contain =), it's likely shorthand
		if strings.Contains(leftSide, ":") ||
			(strings.HasPrefix(rightSide, "/") || !strings.Contains(rightSide, "=")) {
			return true
		}
	}
	return false
}

// Set parses and sets a mount specification
func (m *MountArray) Set(value string) error {
	if value == "" {
		return nil
	}

	// Check for invalid shorthand attempts (colon without equals)
	if strings.Contains(value, ":") && !strings.Contains(value, "=") {
		return fmt.Errorf("invalid mount format: shorthand format requires equals sign")
	}

	currentIsShorthand := isShorthandFormat(value)

	var mount MountSpec
	var err error

	if currentIsShorthand {
		mount, err = parseShorthandMount(value)
	} else {
		mount, err = parseNormalMount(value)
	}

	if err != nil {
		return err
	}

	// Validate path traversal
	if err := validatePath(mount.Path); err != nil {
		return err
	}

	m.Mounts = append(m.Mounts, mount)
	return nil
}

// parseNormalMount parses normal format: name=resource,path=/path,key=key,ro=true
func parseNormalMount(value string) (MountSpec, error) {
	mount := MountSpec{}

	// Handle backward compatibility - if it's just a name, return it
	if !strings.Contains(value, "=") {
		mount.Name = value
		return mount, nil
	}

	fields := strings.Split(value, ",")
	for _, field := range fields {
		parts := strings.SplitN(field, "=", 2)
		if len(parts) != 2 {
			return mount, fmt.Errorf("invalid mount format: field '%s' should be in format key=value", field)
		}

		key, val := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
		switch key {
		case "name":
			mount.Name = val
		case "path":
			mount.Path = val
		case "key":
			mount.Key = val
		case "ro":
			mount.ReadOnly = val == "true"
		default:
			return mount, fmt.Errorf("unknown mount field: '%s', supported fields are 'name', 'path', 'key', 'ro'", key)
		}
	}

	if mount.Name == "" {
		return mount, fmt.Errorf("invalid mount format: name is required")
	}

	// Path is required for new format (when any key=value is specified)
	if mount.Path == "" && (strings.Contains(value, "path=") || strings.Contains(value, "key=") || strings.Contains(value, "ro=")) {
		return mount, fmt.Errorf("mount path is required")
	}

	return mount, nil
}

// parseShorthandMount parses shorthand format: name:key=path:ro or name=path:ro
func parseShorthandMount(value string) (MountSpec, error) {
	mount := MountSpec{}

	// Split by first = to separate name[:key] from path[:ro]
	eqParts := strings.SplitN(value, "=", 2)
	if len(eqParts) != 2 {
		return mount, fmt.Errorf("invalid mount format: shorthand format should be 'name:key=path:ro' or 'name=path:ro'")
	}

	nameKeyPart := eqParts[0]
	pathRoPart := eqParts[1]

	// Parse name[:key]
	nameParts := strings.Split(nameKeyPart, ":")
	mount.Name = nameParts[0]
	if len(nameParts) > 1 {
		mount.Key = nameParts[1]
	}

	// Parse path[:ro]
	pathParts := strings.Split(pathRoPart, ":")
	mount.Path = pathParts[0]
	if len(pathParts) > 1 && pathParts[1] == "ro" {
		mount.ReadOnly = true
	}

	if mount.Name == "" {
		return mount, fmt.Errorf("invalid mount format: name is required")
	}
	if mount.Path == "" {
		return mount, fmt.Errorf("mount path is required")
	}

	return mount, nil
}

// validatePath checks for path traversal attempts
func validatePath(path string) error {
	if path == "" {
		return nil // Empty path is valid for backward compatibility
	}

	// Check for path traversal patterns before cleaning
	if strings.Contains(path, "..") {
		return fmt.Errorf("path traversal is not allowed")
	}

	return nil
}

// ServiceBindingSecretArray holds service binding secret specifications
type ServiceBindingSecretArray struct {
	Names []string
}

// String returns the string representation
func (s *ServiceBindingSecretArray) String() string {
	return strings.Join(s.Names, ",")
}

// Type returns the type name
func (s *ServiceBindingSecretArray) Type() string {
	return "service-binding-secret"
}

// Set adds a service binding secret name
func (s *ServiceBindingSecretArray) Set(value string) error {
	if value == "" {
		return nil
	}
	s.Names = append(s.Names, value)
	return nil
}
