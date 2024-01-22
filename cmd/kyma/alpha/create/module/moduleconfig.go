package module

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/kyma-project/lifecycle-manager/api/v1beta2"

	"github.com/kyma-project/cli/pkg/module"

	"github.com/Masterminds/semver/v3"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Name              string                     `yaml:"name" comment:"required, the name of the Module"`
	Version           string                     `yaml:"version" comment:"required, the version of the Module"`
	Channel           string                     `yaml:"channel" comment:"required, channel that should be used in the ModuleTemplate"`
	ManifestPath      string                     `yaml:"manifest" comment:"required, relative path or remote URL to the manifests"`
	Mandatory         bool                       `yaml:"mandatory" comment:"optional, default=false, indicates whether the module is mandatory to be installed on all clusters"`
	DefaultCRPath     string                     `yaml:"defaultCR" comment:"optional, relative path or remote URL to a YAML file containing the default CR for the module"`
	ResourceName      string                     `yaml:"resourceName" comment:"optional, default={NAME}-{CHANNEL}, the name for the ModuleTemplate that will be created"`
	Namespace         string                     `yaml:"namespace" comment:"optional, default=kcp-system, the namespace where the ModuleTemplate will be deployed"`
	Security          string                     `yaml:"security" comment:"optional, name of the security scanners config file"`
	Internal          bool                       `yaml:"internal" comment:"optional, default=false, determines whether the ModuleTemplate should have the internal flag or not"`
	Beta              bool                       `yaml:"beta" comment:"optional, default=false, determines whether the ModuleTemplate should have the beta flag or not"`
	Labels            map[string]string          `yaml:"labels" comment:"optional, additional labels for the ModuleTemplate"`
	Annotations       map[string]string          `yaml:"annotations" comment:"optional, additional annotations for the ModuleTemplate"`
	CustomStateChecks []v1beta2.CustomStateCheck `yaml:"customStateCheck" comment:"optional, specifies custom state check for module"`
}

const (
	// taken from "github.com/open-component-model/ocm/resources/component-descriptor-v2-schema.yaml"
	moduleNamePattern = "^[a-z][-a-z0-9]*([.][a-z][-a-z0-9]*)*[.][a-z]{2,}(/[a-z][-a-z0-9_]*([.][a-z][-a-z0-9_]*)*)+$"
	namespacePattern  = "^[a-z0-9]+(?:-[a-z0-9]+)*$"
	moduleNameMaxLen  = 255
	namespaceMaxLen   = 253
)

func ParseConfig(filePath string) (*Config, error) {
	data, err := os.ReadFile(filePath)

	if err != nil {
		return nil, fmt.Errorf("error reading module config file %q: %w", filePath, err)
	}

	res := Config{}
	err = yaml.Unmarshal(data, &res)
	if err != nil {
		return nil, fmt.Errorf("error parsing module config file %q: %w", filePath, err)
	}

	return &res, nil
}

func (c *Config) Validate() error {
	return newConfigValidator(c).
		validateName().
		validateNamespace().
		validateVersion().
		validateChannel().
		validateCustomStateChecks().
		do()
}

type configValidationFunc func() error

type configValidator struct {
	config     *Config
	validators []configValidationFunc
}

func newConfigValidator(cnf *Config) *configValidator {
	return &configValidator{
		config:     cnf,
		validators: []configValidationFunc{},
	}
}

func (cv *configValidator) addValidator(fn configValidationFunc) *configValidator {
	cv.validators = append(cv.validators, fn)
	return cv
}

func (cv *configValidator) validateName() *configValidator {
	fn := func() error {
		if len(cv.config.Name) == 0 {
			return fmt.Errorf("%w, module name cannot be empty", ErrNameValidation)
		}
		if len(cv.config.Name) > moduleNameMaxLen {
			return fmt.Errorf("%w, module name length cannot exceed 255 characters", ErrNameValidation)
		}
		matched, err := regexp.MatchString(moduleNamePattern, cv.config.Name)
		if err != nil {
			return fmt.Errorf("failed to evaluate regex for module name pattern: %w", err)
		}
		if !matched {
			return fmt.Errorf("%w for input %q, name must match the required pattern, e.g: 'github.com/path-to/your-repo'",
				ErrNameValidation, cv.config.Name)
		}

		return nil
	}

	return cv.addValidator(fn)
}

func (cv *configValidator) validateNamespace() *configValidator {
	fn := func() error {
		if len(cv.config.Namespace) == 0 {
			return nil
		}
		if len(cv.config.Namespace) > namespaceMaxLen {
			return fmt.Errorf("%w, module name length cannot exceed 253 characters", ErrNamespaceValidation)
		}
		matched, err := regexp.MatchString(namespacePattern, cv.config.Namespace)
		if err != nil {
			return fmt.Errorf("failed to evaluate regex for module namespace pattern: %w", err)
		}
		if !matched {
			return fmt.Errorf("%w for input %q, namespace must contain only small alphanumeric characters and hyphens",
				ErrNamespaceValidation, cv.config.Namespace)
		}

		return nil
	}

	return cv.addValidator(fn)
}

func (cv *configValidator) validateVersion() *configValidator {
	fn := func() error {

		prefix := ""
		val := strings.TrimSpace(cv.config.Version)

		// strip the leading "v", if any, because it's not part of a proper semver
		if strings.HasPrefix(val, "v") {
			prefix = "v"
			val = val[1:]
		}

		sv, err := semver.StrictNewVersion(val)
		if err != nil {
			return fmt.Errorf("%w for input %q, %w", ErrVersionValidation, cv.config.Version, err)
		}

		// restore "v" prefix, if any
		correct := prefix + sv.String()

		if correct != cv.config.Version {
			return fmt.Errorf("%w for input %q, try with %q", ErrVersionValidation, cv.config.Version, correct)
		}
		return nil
	}

	return cv.addValidator(fn)
}

func (cv *configValidator) validateChannel() *configValidator {
	fn := func() error {
		if len(cv.config.Channel) < ChannelMinLength || len(cv.config.Channel) > ChannelMaxLength {
			return fmt.Errorf(
				"%w for input %q, invalid channel length, length should between %d and %d",
				ErrChannelValidation, cv.config.Channel, ChannelMinLength, ChannelMaxLength)
		}
		matched, err := regexp.MatchString(`^[a-z]+$`, cv.config.Channel)
		if err != nil {
			return fmt.Errorf("failed to evaluate regex for channel: %w", err)
		}
		if !matched {
			return fmt.Errorf("%w for input %q, invalid channel format, only allow characters from a-z",
				ErrChannelValidation, cv.config.Channel)
		}
		return nil

	}

	return cv.addValidator(fn)
}

func (cv *configValidator) validateCustomStateChecks() *configValidator {
	fn := func() error {
		cscs := cv.config.CustomStateChecks
		if len(cscs) == 0 {
			return nil
		}
		for _, check := range cscs {
			if len(check.JSONPath) == 0 || len(check.Value) == 0 || len(check.MappedState) == 0 {
				return fmt.Errorf("%w for check %v, not all fields were provided",
					module.ErrCustomStateCheckValidation, check)
			}
			if !module.IsValidMappedState(string(check.MappedState)) {
				return fmt.Errorf("%w because %s is not a valid state name in kyma",
					module.ErrCustomStateCheckValidation, check.MappedState)
			}
		}

		if !module.ContainsAllRequiredStates(cscs) {
			return fmt.Errorf("%w: customStateCheck must contain both required states 'Error' and 'Ready'",
				module.ErrCustomStateCheckValidation)
		}

		return nil

	}

	return cv.addValidator(fn)
}

func (cv *configValidator) do() error {
	for _, v := range cv.validators {
		if err := v(); err != nil {
			return err
		}
	}
	return nil
}
