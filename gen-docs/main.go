package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmd"
)

func main() {
	command, clierr := cmd.NewKymaCMD()
	clierror.Check(clierr)

	docsTargetDir := "./docs/user/gen-docs"
	err := genMarkdownTree(command, docsTargetDir)
	if err != nil {
		fmt.Println("unable to generate docs", err.Error())
		os.Exit(1)
	}

	fmt.Println("Docs successfully generated to the following dir", docsTargetDir)
	os.Exit(0)
}

// does the same as doc.GenMarkdownTree from the "github.com/spf13/cobra/doc" package
// but in the kyma-project.io format
// most of the code is copied the package
func genMarkdownTree(cmd *cobra.Command, dir string) error {
	// gen file for the root command
	basename := getNewMarkdownPrefix() + strings.Replace(cmd.CommandPath(), " ", "_", -1) + ".md"
	filename := filepath.Join(dir, basename)
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	// gen files for all sub-commands
	for _, c := range cmd.Commands() {
		if !c.IsAvailableCommand() || c.IsAdditionalHelpTopicCommand() {
			continue
		}
		if err := genMarkdownTree(c, dir); err != nil {
			return err
		}
	}

	return genMarkdown(cmd, f)
}

func genMarkdown(cmd *cobra.Command, w io.Writer) error {
	cmd.InitDefaultHelpCmd()
	cmd.InitDefaultHelpFlag()

	buf := new(bytes.Buffer)

	printShort(buf, cmd)
	printSynopsis(buf, cmd)

	if cmd.Runnable() {
		buf.WriteString(fmt.Sprintf("```bash\n%s\n```\n\n", cmd.UseLine()))
	}

	printAvailableCommands(buf, cmd)

	if len(cmd.Example) > 0 {
		buf.WriteString("## Examples\n\n")
		buf.WriteString(fmt.Sprintf("```bash\n%s\n```\n\n", cmd.Example))
	}

	if err := printOptions(buf, cmd); err != nil {
		return err
	}

	printSeeAlso(buf, cmd)

	_, err := buf.WriteTo(w)
	return err
}

var (
	fileOrderNumber = 1
)

func getNewMarkdownPrefix() string {
	prefix := fmt.Sprintf("01-%d0-", fileOrderNumber)
	fileOrderNumber++
	return prefix
}

func printShort(buf *bytes.Buffer, cmd *cobra.Command) {
	short := cmd.Short

	buf.WriteString("# " + cmd.CommandPath() + "\n\n")
	if short != "" {
		buf.WriteString(short + "\n\n")
	}
}

func printSynopsis(buf *bytes.Buffer, cmd *cobra.Command) {
	short := cmd.Short
	long := cmd.Long
	if len(long) == 0 {
		long = short
	}

	buf.WriteString("## Synopsis\n\n")
	if long != "" {
		buf.WriteString(long + "\n\n")
	}
}

func printAvailableCommands(buf *bytes.Buffer, cmd *cobra.Command) {
	subCommands := cmd.Commands()
	if len(subCommands) == 0 {
		return
	}

	elems := []printElem{}
	maxNameLen := 0
	for i := range subCommands {
		subCmd := subCommands[i]
		subCmdName := subCmd.Name()
		if len(subCmdName) > maxNameLen {
			maxNameLen = len(subCmdName)
		}
		elems = append(elems, printElem{
			name:        subCmdName,
			description: subCmd.Short,
		})
	}

	buf.WriteString("## Available Commands\n\n")
	buf.WriteString("```bash\n")
	for i := range elems {
		separatorLen := maxNameLen - len(elems[i].name)
		buf.WriteString(fmt.Sprintf("  %s%s - %s\n", elems[i].name, strings.Repeat(" ", separatorLen), elems[i].description))
	}
	buf.WriteString("```\n\n")
}

func printOptions(buf *bytes.Buffer, cmd *cobra.Command) error {
	flags := cmd.NonInheritedFlags()
	flags.SetOutput(buf)
	if flags.HasAvailableFlags() {
		buf.WriteString("## Flags\n\n```bash\n")
		flags.PrintDefaults()
		buf.WriteString("```\n\n")
	}

	parentFlags := cmd.InheritedFlags()
	parentFlags.SetOutput(buf)
	if parentFlags.HasAvailableFlags() {
		buf.WriteString("## Flags inherited from parent commands\n\n```bash\n")
		parentFlags.PrintDefaults()
		buf.WriteString("```\n\n")
	}

	return nil
}

type printElem struct {
	name        string
	description string
}

func printSeeAlso(buf *bytes.Buffer, cmd *cobra.Command) {
	if !hasSeeAlso(cmd) {
		return
	}

	elems := []printElem{}
	maxNameLen := 0
	name := cmd.CommandPath()

	buf.WriteString("## See also\n\n")
	if cmd.HasParent() {
		parent := cmd.Parent()
		pname := parent.CommandPath()
		elem := printElem{
			name:        fmt.Sprintf("[%s](%s)", pname, linkHandler(parent)),
			description: parent.Short,
		}
		elems = append(elems, elem)
		maxNameLen = len(elem.name)
		cmd.VisitParents(func(c *cobra.Command) {
			if c.DisableAutoGenTag {
				cmd.DisableAutoGenTag = c.DisableAutoGenTag
			}
		})
	}

	children := cmd.Commands()
	sort.Sort(byName(children))

	for _, child := range children {
		if !child.IsAvailableCommand() || child.IsAdditionalHelpTopicCommand() {
			continue
		}
		cname := name + " " + child.Name()
		elem := printElem{
			name:        fmt.Sprintf("[%s](%s)", cname, linkHandler(child)),
			description: child.Short,
		}
		elems = append(elems, elem)
		if len(elem.name) > maxNameLen {
			maxNameLen = len(elem.name)
		}
	}

	// print see also
	for i := range elems {
		separatorLen := maxNameLen - len(elems[i].name)
		buf.WriteString(fmt.Sprintf("* %s%s - %s\n", elems[i].name, strings.Repeat(" ", separatorLen), elems[i].description))
	}
}

func linkHandler(cmd *cobra.Command) string {
	name := cmd.CommandPath()
	formatted := strings.ReplaceAll(name, " ", "_")
	return fmt.Sprintf("%s.md", formatted)
}

func hasSeeAlso(cmd *cobra.Command) bool {
	if cmd.HasParent() {
		return true
	}
	for _, c := range cmd.Commands() {
		if !c.IsAvailableCommand() || c.IsAdditionalHelpTopicCommand() {
			continue
		}
		return true
	}
	return false
}

type byName []*cobra.Command

func (s byName) Len() int           { return len(s) }
func (s byName) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s byName) Less(i, j int) bool { return s[i].Name() < s[j].Name() }
