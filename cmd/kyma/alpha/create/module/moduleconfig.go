package module

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/blang/semver/v4"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Name          string            `yaml:"name"`         //required, the name of the Module
	Version       string            `yaml:"version"`      //required, the version of the Module
	Channel       string            `yaml:"channel"`      //required, channel that should be used in the ModuleTemplate
	ManifestPath  string            `yaml:"manifest"`     //required, reference to the manifests, must be a relative file name.
	DefaultCRPath string            `yaml:"defaultCR"`    //optional, reference to a YAML file containing the default CR for the module, must be a relative file name.
	ResourceName  string            `yaml:"resourceName"` //optional, default={NAME}-{CHANNEL}, the name for the ModuleTemplate that will be created
	Security      string            `yaml:"security"`     //optional, name of the security scanners config file
	Internal      bool              `yaml:"internal"`     //optional, default=false, determines whether the ModuleTemplate should have the internal flag or not
	Beta          bool              `yaml:"beta"`         //optional, default=false, determines whether the ModuleTemplate should have the beta flag or not
	Labels        map[string]string `yaml:"labels"`       //optional, additional labels for the ModuleTemplate
	Annotations   map[string]string `yaml:"annotations"`  //optional, additional annotations for the ModuleTemplate
}

const (
	//taken from "github.com/open-component-model/ocm/resources/component-descriptor-v2-schema.yaml"
	moduleNamePattern = "^[a-z][-a-z0-9]*([.][a-z][-a-z0-9]*)*[.][a-z]{2,}(/[a-z][-a-z0-9_]*([.][a-z][-a-z0-9_]*)*)+$"
	moduleNameMaxLen  = 255
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
		validateVersion().
		validateChannel().
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
		matched, _ := regexp.MatchString(moduleNamePattern, cv.config.Name)
		if !matched {
			return fmt.Errorf("%w for input %q, name must match the required pattern, e.g: 'github.com/path-to/your-repo'", ErrNameValidation, cv.config.Name)
		}

		return nil
	}

	return cv.addValidator(fn)
}

func (cv *configValidator) validateVersion() *configValidator {
	fn := func() error {

		prefix := ""
		val := strings.TrimSpace(cv.config.Version)

		//strip the leading "v", if any, because it's not part of a proper semver
		if strings.HasPrefix(val, "v") {
			prefix = "v"
			val = val[1:]
		}

		sv, err := semver.Parse(val)
		if err != nil {
			return fmt.Errorf("%w for input %q, %w", ErrVersionValidation, cv.config.Version, err)
		}

		//restore "v" prefix, if any
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
		matched, _ := regexp.MatchString(`^[a-z]+$`, cv.config.Channel)
		if !matched {
			return fmt.Errorf("%w for input %q, invalid channel format, only allow characters from a-z", ErrChannelValidation, cv.config.Channel)
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
