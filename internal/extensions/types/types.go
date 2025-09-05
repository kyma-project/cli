package types

import (
	"fmt"
	"slices"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/extensions/errors"
	"github.com/kyma-project/cli.v3/internal/extensions/parameters"
	"github.com/spf13/cobra"
)

const (
	ExtensionCMLabelKey   = "kyma-cli/extension"
	ExtensionCMLabelValue = "commands"
	ExtensionCMDataKey    = "kyma-commands.yaml"
)

type Action interface {
	Configure(ActionConfig, ActionConfigOverwrites) clierror.Error
	Run(*cobra.Command, []string) clierror.Error
}

// map of allowed action commands in format ID: ACTION
type ActionsMap map[string]Action

type ActionConfig = map[string]interface{}

type ActionConfigOverwrites = map[string]interface{}

type ConfigmapCommandExtension struct {
	ConfigMapName      string
	ConfigMapNamespace string
	Extension          Extension
}

type Metadata struct {
	// name of the command group
	Name string `yaml:"name"`
	// short description of the command group
	Description string `yaml:"description"`
	// long description of the command group
	DescriptionLong string `yaml:"descriptionLong"`
}

func (m *Metadata) Validate() error {
	if m.Name == "" {
		return errors.New("empty name")
	}

	return nil
}

type Args struct {
	// type of the argument and config field
	// TODO: support many args by adding new type, like `stringArray`
	Type parameters.ConfigFieldType `yaml:"type"`
	// mark if args are required to run command
	Optional bool `yaml:"optional"`
}

func (a *Args) Validate() error {
	if !slices.Contains(parameters.ValidTypes, a.Type) {
		return errors.New(fmt.Sprintf("unknown type '%s'", a.Type))
	}

	return nil
}

type Flag struct {
	// type of the flag and config field
	Type parameters.ConfigFieldType `yaml:"type"`
	// name of the flag
	Name string `yaml:"name"`
	// description of the flag
	Description string `yaml:"description"`
	// optional shorthand of the flag
	Shorthand string `yaml:"shorthand"`
	// default value for the flag
	DefaultValue string `yaml:"default"`
	// mark if flag is required
	Required bool `yaml:"required"`
}

func (f *Flag) Validate() error {
	var errs []error
	if f.Name == "" {
		errs = append(errs, errors.New("empty name"))
	}

	if !slices.Contains(parameters.ValidTypes, f.Type) {
		errs = append(errs, errors.Newf("unknown type '%s'", f.Type))
	}

	return errors.JoinWithSeparator(", ", errs...)
}

type Extension struct {
	// metadata (name, descriptions) for the command
	Metadata Metadata `yaml:"metadata"`
	// id of the functionality that cli will run when user use this command
	Action string `yaml:"uses"`
	// flags used to set specific fields in config
	Flags []Flag `yaml:"flags"`
	// args used to set specific fields in config
	Args *Args `yaml:"args"`
	// additional config pass to the command
	Config ActionConfig `yaml:"with"`
	// list of sub commands
	SubCommands []Extension `yaml:"subCommands"`
}

func (e *Extension) Validate(availableActions ActionsMap) error {
	return e.validateWithPath(".", availableActions)
}

func (e *Extension) validateWithPath(path string, availableActions ActionsMap) error {
	var errs []error
	if metaErr := e.Metadata.Validate(); metaErr != nil {
		errs = append(errs, errors.Newf("wrong %smetadata: %s", path, metaErr.Error()))
	}

	if e.Args != nil {
		if argsErr := e.Args.Validate(); argsErr != nil {
			errs = append(errs, errors.Newf("wrong %sargs: %s", path, argsErr.Error()))
		}
	}

	for i := range e.Flags {
		if flagErr := e.Flags[i].Validate(); flagErr != nil {
			errs = append(errs, errors.Newf("wrong %sflags: %s", path, flagErr))
		}
	}

	for i := range e.SubCommands {
		subCmdErr := e.SubCommands[i].validateWithPath(fmt.Sprintf("%ssubCommands[%d].", path, i), availableActions)
		if subCmdErr != nil {
			errs = append(errs, subCmdErr)
		}
	}

	return errors.NewList(errs...)
}
