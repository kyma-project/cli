package scaffold_test

import (
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

var _ = Describe("Create Scaffold Command", Ordered, func() {
	var initialDir string
	var workDir string
	var workDirCleanup func()

	setup := func() {
		var err error
		initialDir, err = os.Getwd()
		Expect(err).To(BeNil())
		workDir, workDirCleanup = resolveWorkingDirectory()
		err = os.Chdir(workDir)
		Expect(err).To(BeNil())

	}
	teardown := func() {
		err := os.Chdir(initialDir)
		Expect(err).To(BeNil())
		workDirCleanup()
		workDir = ""
		initialDir = ""

	}

	Context("Given an empty directory", func() {
		BeforeAll(func() { setup() })
		AfterAll(func() { teardown() })

		var cmd createScaffoldCmd
		It("When `kyma alpha create scaffold` command is invoked without any args", func() {
			cmd = createScaffoldCmd{}
		})

		It("Then the command should succeed", func() {
			Expect(cmd.execute()).To(Succeed())

			By("And two files are generated")
			Expect(filesIn(workDir)).Should(HaveLen(2))

			By("And the manifest file is generated")
			Expect(filesIn(workDir)).Should(ContainElement("manifest.yaml"))

			By("And the module config file is generated")
			Expect(filesIn(workDir)).Should(ContainElement("scaffold-module-config.yaml"))

			By("And the module config contains expected entries")
			actualModConf := moduleConfigFromFile(workDir, "scaffold-module-config.yaml")
			expectedModConf := (&moduleConfigBuilder{}).defaults().get()
			Expect(actualModConf).To(BeEquivalentTo(expectedModConf))
		})
	})

	Context("Given a directory with an existing module configuration file", func() {
		BeforeAll(func() {
			setup()
			Expect(createMarkerFile("scaffold-module-config.yaml")).To(Succeed())
		})
		AfterAll(func() { teardown() })

		var cmd createScaffoldCmd
		It("When `kyma alpha create scaffold` command is invoked without any args", func() {
			cmd = createScaffoldCmd{}
		})
		It("Then the command should fail", func() {
			err := cmd.execute()
			Expect(err).ShouldNot(BeNil())
			Expect(err.Error()).Should(ContainSubstring("module config file already exists"))

			By("And no files should be generated")
			Expect(filesIn(workDir)).Should(HaveLen(1))
			Expect(filesIn(workDir)).Should(ContainElement("scaffold-module-config.yaml"))
			Expect(getMarkerFileData("scaffold-module-config.yaml")).Should(Equal(markerFileData))
		})
	})

	Context("Given a directory with an existing module configuration file", func() {
		BeforeAll(func() {
			setup()
			Expect(createMarkerFile("scaffold-module-config.yaml")).To(Succeed())
		})
		AfterAll(func() { teardown() })

		var cmd createScaffoldCmd
		It("When `kyma alpha create scaffold` command is invoked with --overwrite flag", func() {
			cmd = createScaffoldCmd{
				overwrite: true,
			}
		})

		It("Then the command should succeed", func() {
			Expect(cmd.execute()).To(Succeed())

			By("And two files are generated")
			Expect(filesIn(workDir)).Should(HaveLen(2))

			By("And the manifest file is generated")
			Expect(filesIn(workDir)).Should(ContainElement("manifest.yaml"))

			By("And the module config file is generated")
			Expect(filesIn(workDir)).Should(ContainElement("scaffold-module-config.yaml"))

			By("And the module config contains expected entries")
			actualModConf := moduleConfigFromFile(workDir, "scaffold-module-config.yaml")
			expectedModConf := (&moduleConfigBuilder{}).defaults().get()
			Expect(actualModConf).To(BeEquivalentTo(expectedModConf))
		})
	})

	Context("Given an empty directory", func() {
		BeforeAll(func() { setup() })
		AfterAll(func() { teardown() })

		var cmd createScaffoldCmd
		It("When `kyma alpha create scaffold` command args override defaults", func() {
			cmd = createScaffoldCmd{
				moduleName:                    "github.com/custom/module",
				moduleVersion:                 "3.2.1",
				moduleChannel:                 "custom",
				moduleConfigFileFlag:          "custom-module-config.yaml",
				genManifestFlag:               "custom-manifest.yaml",
				genDefaultCRFlag:              "custom-default-cr.yaml",
				genSecurityScannersConfigFlag: "custom-security-scanners-config.yaml",
			}
		})
		It("Then the command should succeed", func() {
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

			By("And the module config contains expected entries")
			actualModConf := moduleConfigFromFile(workDir, "custom-module-config.yaml")
			expectedModConf := cmd.toConfigBuilder().get()
			Expect(actualModConf).To(BeEquivalentTo(expectedModConf))
		})
	})

	Context("Given a directory with existing files", func() {
		BeforeAll(func() {
			setup()
			Expect(createMarkerFile("custom-manifest.yaml")).To(Succeed())
			Expect(createMarkerFile("custom-default-cr.yaml")).To(Succeed())
			Expect(createMarkerFile("custom-security-scanners-config.yaml")).To(Succeed())

		})
		AfterAll(func() { teardown() })

		var cmd createScaffoldCmd
		It("When `kyma alpha create scaffold` command is invoked with arguments that match existing files names", func() {
			cmd = createScaffoldCmd{
				genManifestFlag:               "custom-manifest.yaml",
				genDefaultCRFlag:              "custom-default-cr.yaml",
				genSecurityScannersConfigFlag: "custom-security-scanners-config.yaml",
			}
		})
		It("Then the command should succeed", func() {
			Expect(cmd.execute()).To(Succeed())

			By("And there should be four files in the directory")
			Expect(filesIn(workDir)).Should(HaveLen(4))

			By("And the manifest file is reused (not generated)")
			Expect(getMarkerFileData("custom-manifest.yaml")).Should(Equal(markerFileData))

			By("And the defaultCR file is reused (not generated)")
			Expect(getMarkerFileData("custom-default-cr.yaml")).Should(Equal(markerFileData))

			By("And the security-scanners-config file is reused (not generated)")
			Expect(getMarkerFileData("custom-security-scanners-config.yaml")).Should(Equal(markerFileData))

			By("And the module config file is generated")
			Expect(filesIn(workDir)).Should(ContainElement("scaffold-module-config.yaml"))

			By("And module config contains expected entries")
			actualModConf := moduleConfigFromFile(workDir, "scaffold-module-config.yaml")
			expectedModConf := cmd.toConfigBuilder().get()
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

func filesIn(dir string) []string {
	fi, err := os.Stat(dir)
	Expect(err).To(BeNil())
	Expect(fi.IsDir()).To(BeTrueBecause("The provided path should be a directory: %s", dir))

	dirFs := os.DirFS(dir)
	entries, err := fs.ReadDir(dirFs, ".")
	Expect(err).To(BeNil())

	res := []string{}
	for _, ent := range entries {
		if ent.Type().IsRegular() {
			res = append(res, ent.Name())
		}
	}

	return res
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

type createScaffoldCmd struct {
	moduleName                    string
	moduleVersion                 string
	moduleChannel                 string
	moduleConfigFileFlag          string
	genDefaultCRFlag              string
	genSecurityScannersConfigFlag string
	genManifestFlag               string
	overwrite                     bool
}

func (cmd *createScaffoldCmd) execute() error {
	var command *exec.Cmd

	args := []string{"alpha", "create", "scaffold"}

	if cmd.moduleName != "" {
		args = append(args, fmt.Sprintf("--module-name=%s", cmd.moduleName))
	}

	if cmd.moduleVersion != "" {
		args = append(args, fmt.Sprintf("--module-version=%s", cmd.moduleVersion))
	}

	if cmd.moduleChannel != "" {
		args = append(args, fmt.Sprintf("--module-channel=%s", cmd.moduleChannel))
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

	if cmd.overwrite {
		args = append(args, "--overwrite=true")
	}

	command = exec.Command("kyma", args...)
	cmdOut, err := command.CombinedOutput()

	if err != nil {
		return fmt.Errorf("create scaffold command failed with output: %s and error: %w", cmdOut, err)
	}
	return nil
}

func (cmd *createScaffoldCmd) toConfigBuilder() *moduleConfigBuilder {
	res := &moduleConfigBuilder{}
	res.defaults()
	if cmd.moduleName != "" {
		res.withName(cmd.moduleName)
	}
	if cmd.moduleVersion != "" {
		res.withVersion(cmd.moduleVersion)
	}
	if cmd.moduleChannel != "" {
		res.withChannel(cmd.moduleChannel)
	}
	if cmd.genDefaultCRFlag != "" {
		res.withDefaultCRPath(cmd.genDefaultCRFlag)
	}
	if cmd.genSecurityScannersConfigFlag != "" {
		res.withSecurityScannersPath(cmd.genSecurityScannersConfigFlag)
	}
	if cmd.genManifestFlag != "" {
		res.withManifestPath(cmd.genManifestFlag)
	}
	return res
}

// moduleConfigBuilder is used to simplify module.Config creation for testing purposes
type moduleConfigBuilder struct {
	module.Config
}

func (mcb *moduleConfigBuilder) get() *module.Config {
	res := mcb.Config
	return &res
}
func (mcb *moduleConfigBuilder) withName(val string) *moduleConfigBuilder {
	mcb.Name = val
	return mcb
}
func (mcb *moduleConfigBuilder) withVersion(val string) *moduleConfigBuilder {
	mcb.Version = val
	return mcb
}
func (mcb *moduleConfigBuilder) withChannel(val string) *moduleConfigBuilder {
	mcb.Channel = val
	return mcb
}
func (mcb *moduleConfigBuilder) withManifestPath(val string) *moduleConfigBuilder {
	mcb.ManifestPath = val
	return mcb
}
func (mcb *moduleConfigBuilder) withDefaultCRPath(val string) *moduleConfigBuilder {
	mcb.DefaultCRPath = val
	return mcb
}
func (mcb *moduleConfigBuilder) withSecurityScannersPath(val string) *moduleConfigBuilder {
	mcb.Security = val
	return mcb
}
func (mcb *moduleConfigBuilder) defaults() *moduleConfigBuilder {
	return mcb.
		withName("kyma-project.io/module/mymodule").
		withVersion("0.0.1").
		withChannel("regular").
		withManifestPath("manifest.yaml")
}
