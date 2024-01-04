package scaffold_test

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path"

	"github.com/kyma-project/cli/cmd/kyma/alpha/create/module"
	"gopkg.in/yaml.v3"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	markerFileData = "test-marker"
)

var _ = Describe("Create Scaffold Command", func() {
	var initialDir string
	var workDir string
	var workDirCleanup func()

	BeforeEach(func() {
		var err error
		workDir, workDirCleanup = resolveWorkingDirectory()
		initialDir, err = os.Getwd()
		Expect(err).To(BeNil())
		err = os.Chdir(workDir)
		Expect(err).To(BeNil())
	})

	AfterEach(func() {

		err := os.Chdir(initialDir)
		Expect(err).To(BeNil())
		if workDirCleanup != nil {
			fmt.Println("foo")
		}
		//workDir = ""
	})

	Context("Given an empty directory", func() {
		It("When `kyma alpha create scaffold` command is executed without any args", func() {
			cmd := CreateScaffoldCmd{}
			Expect(cmd.execute()).To(Succeed())

			By("Then the manifest file is generated")
			Expect(filesIn(workDir)).Should(ContainElement("manifest.yaml"))

			By("And the module config file is generated")
			Expect(filesIn(workDir)).Should(ContainElement("scaffold-module-config.yaml"))

			By("And module config contains expected entries")
			actualModConf, err := moduleConfigFromFile(workDir, "scaffold-module-config.yaml")
			Expect(err).To(BeNil())
			expectedModConf := (&ModuleConfigBuilder{}).defaults().get()
			Expect(actualModConf).To(BeEquivalentTo(expectedModConf))
		})
	})

	Context("Given a directory with existing module configuration file", func() {
		It("When `kyma alpha create scaffold` command is executed", func() {
			Expect(createMarkerFile("scaffold-module-config.yaml")).To(Succeed())

			cmd := CreateScaffoldCmd{}

			By("Then the command should fail")
			err := cmd.execute()
			Expect(err.Error()).Should(ContainSubstring("scaffold module config file already exists"))

			By("And no files should be generated")
			files, err := filesIn(workDir)
			Expect(err).Should(BeNil())
			Expect(files).Should(HaveLen(1))
			Expect(files).Should(ContainElement("scaffold-module-config.yaml"))
			Expect(getMarkerFileData("scaffold-module-config.yaml")).Should(Equal(markerFileData))
		})
	})

	Context("Given a directory with existing module configuration file", func() {
		It("When `kyma alpha create scaffold` command is executed with --overwrite flag", func() {
			Expect(createMarkerFile("scaffold-module-config.yaml")).To(Succeed())

			cmd := CreateScaffoldCmd{
				overwrite: true,
			}

			By("Then the command should succeed")
			Expect(cmd.execute()).To(Succeed())

			By("And the manifest file is generated")
			Expect(filesIn(workDir)).Should(ContainElement("manifest.yaml"))

			By("And the module config file is generated")
			Expect(filesIn(workDir)).Should(ContainElement("scaffold-module-config.yaml"))

			By("And module config contains expected entries")
			actualModConf, err := moduleConfigFromFile(workDir, "scaffold-module-config.yaml")
			Expect(err).To(BeNil())
			expectedModConf := (&ModuleConfigBuilder{}).defaults().get()
			Expect(actualModConf).To(BeEquivalentTo(expectedModConf))
		})
	})

})

func getMarkerFileData(name string) string {
	data, err := os.ReadFile(name)
	Expect(err).To(BeNil())
	return string(data)
}

func createMarkerFile(name string) error {
	err := os.WriteFile(name, []byte(markerFileData), 0600)
	return err
}

func moduleConfigFromFile(dir, fileName string) (*module.Config, error) {
	filePath := path.Join(dir, fileName)
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	res := module.Config{}
	if err = yaml.Unmarshal(data, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

func filesIn(dir string) ([]string, error) {
	fi, err := os.Stat(dir)
	if err != nil {
		return nil, err
	}
	if !fi.IsDir() {
		return nil, errors.New(fmt.Sprintf("Not a directory: %s", dir))
	}
	dirFs := os.DirFS(dir)
	entries, err := fs.ReadDir(dirFs, ".")
	if err != nil {
		return nil, err
	}

	res := []string{}
	for _, ent := range entries {
		if ent.Type().IsRegular() {
			res = append(res, ent.Name())
		}
	}

	return res, nil
}

func resolveWorkingDirectory() (path string, cleanup func()) {
	path = os.Getenv("SCAFFOLD_DIR")
	if len(path) > 0 {
		cleanup = func() {}
	} else {
		var err error
		path, err = os.MkdirTemp("", "create_scaffold_test")
		if err != nil {
			Fail(err.Error())
		}
		fmt.Printf("\nCreated temporary directory: %s\n", path)
		cleanup = func() {
			os.RemoveAll(path)
		}
	}
	return
}

type CreateScaffoldCmd struct {
	overwrite bool
}

func (cmd *CreateScaffoldCmd) execute() error {
	var createScaffoldCmd *exec.Cmd

	args := []string{"alpha", "create", "scaffold"}

	if cmd.overwrite {
		args = append(args, "--overwrite=true")
	}

	createScaffoldCmd = exec.Command("kyma", args...)
	cmdOut, err := createScaffoldCmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf("create scaffold command failed with output: %s and error: %w", cmdOut, err)
	}
	return nil
}

type ModuleConfigBuilder struct {
	module.Config
}

func (mcb *ModuleConfigBuilder) get() *module.Config {
	res := mcb.Config
	return &res
}
func (mcb *ModuleConfigBuilder) withName(val string) *ModuleConfigBuilder {
	mcb.Name = val
	return mcb
}
func (mcb *ModuleConfigBuilder) withVersion(val string) *ModuleConfigBuilder {
	mcb.Version = val
	return mcb
}
func (mcb *ModuleConfigBuilder) withChannel(val string) *ModuleConfigBuilder {
	mcb.Channel = val
	return mcb
}
func (mcb *ModuleConfigBuilder) withManifestPath(val string) *ModuleConfigBuilder {
	mcb.ManifestPath = val
	return mcb
}
func (mcb *ModuleConfigBuilder) defaults() *ModuleConfigBuilder {
	return mcb.
		withName("kyma-project.io/module/mymodule").
		withVersion("0.0.1").
		withChannel("regular").
		withManifestPath("manifest.yaml")
}
