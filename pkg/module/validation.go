package module

import (
	"errors"
	"fmt"
	"strings"

	amv "k8s.io/apimachinery/pkg/util/validation"
)

// ValidateName checks if the name is at least three characters long and if it conforms to the "RFC 1035 Label Names" specification (K8s compatibility requirement)
func ValidateName(name string) error {
	if len(name) < 3 {
		return errors.New("Invalid module name: name must be at least three characters long")
	}

	violations := amv.IsDNS1035Label(name)
	if len(violations) == 1 {
		return fmt.Errorf("Invalid module name: %s", violations[0])
	}
	if len(violations) > 1 {
		vl := "\n - " + strings.Join(violations, "\n - ")
		return fmt.Errorf("Invalid module name: %s", vl)
	}

	return nil
}
