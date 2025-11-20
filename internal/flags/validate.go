package flags

import (
	"errors"
	"fmt"
	"strings"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/spf13/pflag"
)

type Rule func(flags *pflag.FlagSet) error

// validates flags based on given rules
// returns nil if validation pass or clierror.Error containing all failed validation rules as hints
func Validate(flags *pflag.FlagSet, rules ...Rule) clierror.Error {
	validationErrors := []error{}
	for _, rule := range rules {
		err := rule(flags)
		if err != nil {
			validationErrors = append(validationErrors, err)
		}
	}

	if len(validationErrors) != 0 {
		return clierror.New("failed to validate given flags", errorHints(validationErrors...)...)
	}

	return nil
}

// same as cobra.MarkFlagRequired but with better error msg
func MarkRequired(flags ...string) Rule {
	return func(flagSet *pflag.FlagSet) error {
		missingFlags := listMissingChangedFlags(flagSet, flags...)
		if len(missingFlags) > 0 {
			return fmt.Errorf("all flags in group [%s] must be set, missing [%s]",
				strings.Join(flags, " "), strings.Join(missingFlags, " "))
		}

		return nil
	}
}

// same as cobra.MarkFlagsRequiredTogether but with better error msg
func MarkRequiredTogether(flags ...string) Rule {
	return func(flagSet *pflag.FlagSet) error {
		if !anyOfFlagsChanges(flagSet, flags...) {
			// ignore rule because zero from given flags are set
			return nil
		}

		// return MarkRequired(flags...)(flagSet)
		missingFlags := listMissingChangedFlags(flagSet, flags...)
		if len(missingFlags) > 0 {
			// not all flags are used
			return fmt.Errorf("all flags in group [%s] must be set if any is used, missing [%s]",
				strings.Join(flags, " "), strings.Join(missingFlags, " "))
		}

		return nil
	}
}

// same as cobra.MarkFlagsOneRequired but with better error msg
func MarkOneRequired(flags ...string) Rule {
	return func(flagSet *pflag.FlagSet) error {
		if !anyOfFlagsChanges(flagSet, flags...) {
			return fmt.Errorf("at least one of the flags from the group [%s] must be used", strings.Join(flags, " "))
		}

		return nil
	}
}

// expect exactly one of flags to be used
func MarkExactlyOneRequired(flags ...string) Rule {
	return func(flagSet *pflag.FlagSet) error {
		usedFlags := listChangedFlags(flagSet, flags...)
		if len(usedFlags) > 1 {
			return fmt.Errorf("exactly one from group [%s] must be set, used [%s]",
				strings.Join(flags, " "), strings.Join(usedFlags, " "))
		} else if len(usedFlags) == 0 {
			return fmt.Errorf("exactly one from group [%s] must be set", strings.Join(flags, " "))
		}

		return nil
	}
}

// expect prerequisiteFlags to be used if flag is used
func MarkPrerequisites(flag string, prerequisiteFlags ...string) Rule {
	return func(flagSet *pflag.FlagSet) error {
		if !anyOfFlagsChanges(flagSet, flag) {
			// flag is not used
			return nil
		}

		missingFlags := listMissingChangedFlags(flagSet, prerequisiteFlags...)
		if len(missingFlags) > 0 {
			// not all prerequisiteFlags are used
			return fmt.Errorf("all flags in group [%s] must be set when [%s] flag is used, missing [%s]",
				strings.Join(prerequisiteFlags, " "), flag, strings.Join(missingFlags, " "))
		}

		return nil
	}
}

func MarkUnsupported(flag string, message string) Rule {
	return func(flagSet *pflag.FlagSet) error {
		if anyOfFlagsChanges(flagSet, flag) {
			return errors.New(message)
		}
		return nil
	}
}

// expect exclusiveFlags to be not used if flag is used
func MarkExclusive(flag string, exclusiveFlags ...string) Rule {
	return func(flagSet *pflag.FlagSet) error {
		if !anyOfFlagsChanges(flagSet, flag) {
			// flag is not used
			return nil
		}

		usedFlags := listChangedFlags(flagSet, exclusiveFlags...)
		if len(usedFlags) > 0 {
			return fmt.Errorf("flags in group [%s] can't be used together with [%s], used [%s]",
				strings.Join(exclusiveFlags, " "), flag, strings.Join(usedFlags, " "))
		}

		return nil
	}
}

// same as cobra.MarkFlagsMutuallyExclusive but with better error msg
func MarkMutuallyExclusive(flags ...string) Rule {
	return func(flagSet *pflag.FlagSet) error {
		usedFlags := []string{}
		flagSet.VisitAll(func(f *pflag.Flag) {
			if isOneOf(f.Name, flags...) && f.Changed {
				usedFlags = append(usedFlags, f.Name)
			}
		})

		if len(listChangedFlags(flagSet, flags...)) > 1 {
			return fmt.Errorf("only one flag from group [%s] can be used at the same time, used [%s]",
				strings.Join(flags, " "), strings.Join(usedFlags, " "))
		}

		return nil
	}
}

func listChangedFlags(flagSet *pflag.FlagSet, fromFlags ...string) []string {
	usedFlags := []string{}
	flagSet.VisitAll(func(f *pflag.Flag) {
		if isOneOf(f.Name, fromFlags...) && f.Changed {
			usedFlags = append(usedFlags, f.Name)
		}
	})

	return usedFlags
}

func listMissingChangedFlags(flagSet *pflag.FlagSet, fromFlags ...string) []string {
	missingFlags := []string{}
	flagSet.VisitAll(func(f *pflag.Flag) {
		if isOneOf(f.Name, fromFlags...) && !f.Changed {
			missingFlags = append(missingFlags, f.Name)
		}
	})

	return missingFlags
}

func anyOfFlagsChanges(flagSet *pflag.FlagSet, flags ...string) bool {
	isOneChanged := false
	flagSet.VisitAll(func(f *pflag.Flag) {
		if isOneOf(f.Name, flags...) && f.Changed {
			isOneChanged = true
		}
	})

	return isOneChanged
}

func isOneOf(key string, values ...string) bool {
	for _, value := range values {
		if key == value {
			return true
		}
	}

	return false
}

func errorHints(errors ...error) []string {
	errMsgs := make([]string, len(errors))
	for i := range errors {
		errMsgs[i] = errors[i].Error()
	}

	return errMsgs
}
