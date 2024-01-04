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
		initialDir, err = os.Getwd()
		Expect(err).To(BeNil())
		workDir, workDirCleanup = resolveWorkingDirectory()
		err = os.Chdir(workDir)
		Expect(err).To(BeNil())
	})

	AfterEach(func() {
		err := os.Chdir(initialDir)
		Expect(err).To(BeNil())
		workDirCleanup()
		workDir = ""
		initialDir = ""
	})

	Context("Given an empty directory", func() {
		It("When `kyma alpha create scaffold` command is executed without any args", func() {
			cmd := CreateScaffoldCmd{}
			Expect(cmd.execute()).To(Succeed())

			By("Then two files are generated")
			Expect(filesIn(workDir)).Should(HaveLen(2))

			By("And the manifest file is generated")
			Expect(filesIn(workDir)).Should(ContainElement("manifest.yaml"))

			By("And the module config file is generated")
			Expect(filesIn(workDir)).Should(ContainElement("scaffold-module-config.yaml"))

			By("And module config contains expected entries")
			actualModConf := moduleConfigFromFile(workDir, "scaffold-module-config.yaml")
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

			By("And two files are generated")
			Expect(filesIn(workDir)).Should(HaveLen(2))

			By("And the manifest file is generated")
			Expect(filesIn(workDir)).Should(ContainElement("manifest.yaml"))

			By("And the module config file is generated")
			Expect(filesIn(workDir)).Should(ContainElement("scaffold-module-config.yaml"))

			By("And module config contains expected entries")
			actualModConf := moduleConfigFromFile(workDir, "scaffold-module-config.yaml")
			expectedModConf := (&ModuleConfigBuilder{}).defaults().get()
			Expect(actualModConf).To(BeEquivalentTo(expectedModConf))
		})
	})

	Context("Given an empty directory", func() {
		It("When `kyma alpha create scaffold` command args override default names", func() {
			cmd := CreateScaffoldCmd{
				moduleConfigFileFlag:          "custom-module-config.yaml",
				genManifestFlag:               "custom-manifest.yaml",
				genDefaultCRFlag:              "custom-default-cr.yaml",
				genSecurityScannersConfigFlag: "custom-security-scanners-config.yaml",
			}

			By("Then the command should succeed")
			Expect(cmd.execute()).To(Succeed())

			By("And four files are generated")
			Expect(filesIn(workDir)).Should(HaveLen(4))

			By("And the manifest file is generated")
			Expect(filesIn(workDir)).Should(ContainElement("custom-manifest.yaml"))

			By("And the defaultCR file is generated")
			Expect(filesIn(workDir)).Should(ContainElement("custom-default-cr.yaml"))

			By("And the security-scanners-config file is generated")
			Expect(filesIn(workDir)).Should(ContainElement("custom-security-scanners-config.yaml"))

			By("And the module config file is generated")
			Expect(filesIn(workDir)).Should(ContainElement("custom-module-config.yaml"))

			By("And module config contains expected entries")
			actualModConf := moduleConfigFromFile(workDir, "custom-module-config.yaml")
			expectedModConf := (&ModuleConfigBuilder{}).
				defaults().
				withManifestPath("custom-manifest.yaml").
				withDefaultCRPath("custom-default-cr.yaml").
				withSecurityScannersPath("custom-security-scanners-config.yaml").
				get()
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

func moduleConfigFromFile(dir, fileName string) *module.Config {
	filePath := path.Join(dir, fileName)
	data, err := os.ReadFile(filePath)
	Expect(err).To(BeNil())
	res := module.Config{}
	err = yaml.Unmarshal(data, &res)
	Expect(err).To(BeNil())
	return &res
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
		cleanup = func() {
			os.RemoveAll(path)
		}
	}
	return
}

type CreateScaffoldCmd struct {
	overwrite                     bool
	moduleConfigFileFlag          string
	genDefaultCRFlag              string
	genSecurityScannersConfigFlag string
	genManifestFlag               string
}

func (cmd *CreateScaffoldCmd) execute() error {
	var createScaffoldCmd *exec.Cmd

	args := []string{"alpha", "create", "scaffold"}

	if cmd.overwrite {
		args = append(args, "--overwrite=true")
	}

	if cmd.moduleConfigFileFlag != "" {
		args = append(args, fmt.Sprintf("--module-config=%s", cmd.moduleConfigFileFlag))
	}

	if cmd.genDefaultCRFlag != "" {
		args = append(args, fmt.Sprintf("--gen-default-cr=%s", cmd.genDefaultCRFlag))
	}

	if cmd.genSecurityScannersConfigFlag != "" {
		args = append(args, fmt.Sprintf("--gen-security-config=%s", cmd.genSecurityScannersConfigFlag))
	}

	if cmd.genManifestFlag != "" {
		args = append(args, fmt.Sprintf("--gen-manifest=%s", cmd.genManifestFlag))
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
func (mcb *ModuleConfigBuilder) withDefaultCRPath(val string) *ModuleConfigBuilder {
	mcb.DefaultCRPath = val
	return mcb
}
func (mcb *ModuleConfigBuilder) withSecurityScannersPath(val string) *ModuleConfigBuilder {
	mcb.Security = val
	return mcb
}
func (mcb *ModuleConfigBuilder) defaults() *ModuleConfigBuilder {
	return mcb.
		withName("kyma-project.io/module/mymodule").
		withVersion("0.0.1").
		withChannel("regular").
		withManifestPath("manifest.yaml")
}
