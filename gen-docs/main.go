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
	"github.com/spf13/pflag"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmd"
)

func main() {
	command, clierr := cmd.NewKymaCMD()
	clierror.Check(clierr)

	docsTargetDir := "./docs/user/gen-docs"

	command.InitDefaultCompletionCmd()
	command.InitDefaultHelpCmd()

	err := genMarkdownTree(command, docsTargetDir)
	if err != nil {
		fmt.Println("unable to generate docs", err.Error())
		os.Exit(1)
	}

	err = genSidebarTree(command, docsTargetDir)
	if err != nil {
		fmt.Println("unable to create _sidebar.md file", err.Error())
		os.Exit(1)
	}

	err = genReadme(docsTargetDir)
	if err != nil {
		fmt.Println("unable to create README.md file", err.Error())
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
	basename := cmdFileBasename(cmd)
	filename := filepath.Join(dir, basename)
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	// gen files for all sub-commands
	for _, c := range cmd.Commands() {
		if err := genMarkdownTree(c, dir); err != nil {
			return err
		}
	}

	return genMarkdown(cmd, f)
}

// generates the _sidebar.md file that orders .md files with documentation on the dashboard
func genSidebarTree(cmd *cobra.Command, dir string) error {
	buf := bytes.NewBuffer([]byte{})

	sidebarFile, err := os.Create(filepath.Join(dir, "_sidebar.md"))
	if err != nil {
		return err
	}
	defer sidebarFile.Close()

	buf.WriteString("<!-- markdown-link-check-disable -->\n")
	buf.WriteString("* [Back to Kyma CLI](/cli/user/README.md)\n")
	genSidebar(cmd, buf, 0)
	buf.WriteString("<!-- markdown-link-check-enable -->")

	_, err = buf.WriteTo(sidebarFile)
	return err
}

// generates README.md in the gen-docs dir that helps with rendering of the _sidebar for the kyma-project.io
func genReadme(dir string) error {
	buf := bytes.NewBuffer([]byte{})

	f, err := os.Create(filepath.Join(dir, "README.md"))
	if err != nil {
		return err
	}
	defer f.Close()

	buf.WriteString("# Commands\n\n")
	buf.WriteString("In this section, you can find the available Kyma CLI commands.\n")

	_, err = buf.WriteTo(f)
	return err
}

func genMarkdown(cmd *cobra.Command, w io.Writer) error {
	buf := new(bytes.Buffer)

	printShort(buf, cmd)
	printSynopsis(buf, cmd)
	printAvailableCommands(buf, cmd)
	printExamples(buf, cmd)
	printFlags(buf, cmd)
	printSeeAlso(buf, cmd)

	_, err := buf.WriteTo(w)
	return err
}

func genSidebar(cmd *cobra.Command, buf *bytes.Buffer, indentMultiplier int) {
	buf.WriteString(fmt.Sprintf("* [%s](/cli/user/gen-docs/%s)\n", cmd.CommandPath(), cmdFileBasename(cmd)))

	for _, subCmd := range cmd.Commands() {
		genSidebar(subCmd, buf, indentMultiplier+1)
	}
}

func printExamples(buf *bytes.Buffer, cmd *cobra.Command) {
	if len(cmd.Example) == 0 {
		return
	}

	buf.WriteString("## Examples\n\n")
	buf.WriteString(fmt.Sprintf("```bash\n%s\n```\n\n", cmd.Example))
}

func printShort(buf *bytes.Buffer, cmd *cobra.Command) {
	short := cmd.Short

	buf.WriteString("# " + cmd.CommandPath() + "\n\n")

	// add '.' at the end of the description
	if short[len(short)-1] != '.' {
		short += "."
	}

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

	buf.WriteString(fmt.Sprintf("```bash\n%s\n```\n\n", cmd.UseLine()))
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
	buf.WriteString("```text\n")
	for i := range elems {
		separatorLen := maxNameLen - len(elems[i].name)
		buf.WriteString(fmt.Sprintf("  %s%s - %s\n", elems[i].name, strings.Repeat(" ", separatorLen), elems[i].description))
	}
	buf.WriteString("```\n\n")
}

type printElem struct {
	name        string
	description string
}

func printFlags(buf *bytes.Buffer, cmd *cobra.Command) {
	flags := cmd.NonInheritedFlags()
	parentFlags := cmd.InheritedFlags()

	if !flags.HasAvailableFlags() && !parentFlags.HasAvailableFlags() {
		return
	}

	elems := []printElem{}
	maxNameLen := 0

	// collect flags
	flags.VisitAll(func(f *pflag.Flag) {
		elem := getFlagPrinElem(f)
		elems = append(elems, elem)
		if len(elem.name) > maxNameLen {
			maxNameLen = len(elem.name)
		}
	})

	// collect parent flags
	parentFlags.VisitAll(func(f *pflag.Flag) {
		elem := getFlagPrinElem(f)
		elems = append(elems, elem)
		if len(elem.name) > maxNameLen {
			maxNameLen = len(elem.name)
		}
	})

	// print flags
	buf.WriteString("## Flags\n\n```text\n")
	for _, elem := range elems {
		// flag section with shorthand and separators
		nameSection := fmt.Sprintf("  %s%s   ", elem.name, strings.Repeat(" ", maxNameLen-len(elem.name)))
		// description section with optional multilines
		descriptionSection := strings.Replace(elem.description, "\n", "\n"+strings.Repeat(" ", len(nameSection)), -1)
		// print
		buf.WriteString(fmt.Sprintf("%s%s\n", nameSection, descriptionSection))
	}
	buf.WriteString("```\n\n")
}

func getFlagPrinElem(f *pflag.Flag) printElem {
	shorthandSection := "    "
	if f.Shorthand != "" {
		shorthandSection = fmt.Sprintf("-%s, ", f.Shorthand)
	}

	typeSection := ""
	if f.Value.Type() != "bool" {
		typeSection = fmt.Sprintf(" %s", f.Value.Type())
	}

	descriptionSection := f.Usage
	if f.DefValue != "" && f.Value.Type() != "bool" {
		descriptionSection += fmt.Sprintf(" (default \"%s\")", f.DefValue)
	}

	return printElem{
		name:        fmt.Sprintf("%s--%s%s", shorthandSection, f.Name, typeSection),
		description: descriptionSection,
	}
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

func cmdFileBasename(cmd *cobra.Command) string {
	return strings.Replace(cmd.CommandPath(), " ", "_", -1) + ".md"
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
