package types

import (
	"errors"
	"fmt"
	"slices"

	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/extensions/parameters"
	"github.com/spf13/cobra"
)

const (
	ExtensionCMLabelKey   = "kyma-cli/extension"
	ExtensionCMLabelValue = "commands"
	ExtensionCMDataKey    = "kyma-commands.yaml"
)

// represents the *cobra.Command{}.Run() func type
type CmdRun func(*cobra.Command, []string)

// map of allowed action commands in format ID: FUNC
type ActionsMap map[string]func(*cmdcommon.KymaConfig, ActionConfig) CmdRun

type ActionConfig = map[string]interface{}

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
	// path to the config fild that will be updated with value from args
	ConfigPath string `yaml:"configPath"`
}

func (a *Args) Validate() error {
	var err error
	if !slices.Contains(parameters.ValidTypes, a.Type) {
		err = errors.Join(err, fmt.Errorf("unknown type '%s'", a.Type))
	}

	if a.ConfigPath == "" {
		err = errors.Join(err, errors.New("empty ConfigPath"))
	}

	return err
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
	// path to the config fild that will be updated with value from the flag
	ConfigPath string `yaml:"configPath"`
	// default value for the flag
	DefaultValue string `yaml:"default"`
	// mark if flag is required
	Required bool `yaml:"required"`
}

func (f *Flag) Validate() error {
	var err error
	if !slices.Contains(parameters.ValidTypes, f.Type) {
		err = errors.Join(err, fmt.Errorf("unknown type '%s'", f.Type))
	}

	if f.ConfigPath == "" {
		err = errors.Join(err, errors.New("empty ConfigPath"))
	}

	return err
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

func (e *Extension) Default() {
	if e.Config == nil {
		// default action config to empty (not nil) value
		e.Config = ActionConfig{}
	}

	for i := range e.SubCommands {
		e.SubCommands[i].Default()
	}
}

func (e *Extension) Validate(availableActions ActionsMap) error {
	return e.validateWithPath(".", availableActions)
}

func (e *Extension) validateWithPath(path string, availableActions ActionsMap) error {
	var err error
	if _, ok := availableActions[e.Action]; e.Action != "" && !ok {
		err = errors.Join(err, fmt.Errorf("wrong %suses: unsupported value '%s'", path, e.Action))
	}

	if metaErr := e.Metadata.Validate(); metaErr != nil {
		err = errors.Join(err, fmt.Errorf("wrong %smetadata: %s", path, metaErr.Error()))
	}

	if e.Args != nil {
		if argsErr := e.Args.Validate(); argsErr != nil {
			err = errors.Join(err, fmt.Errorf("wrong %sargs: %s", path, argsErr.Error()))
		}
	}

	for i := range e.Flags {
		if flagErr := e.Flags[i].Validate(); flagErr != nil {
			err = errors.Join(err, fmt.Errorf("wrong %sflags: %s", path, flagErr))
		}
	}

	for i := range e.SubCommands {
		subCmdErr := e.SubCommands[i].validateWithPath(fmt.Sprintf("%ssubCommands[%d].", path, i), availableActions)
		if subCmdErr != nil {
			err = errors.Join(err, subCmdErr)
		}
	}

	return err
}
