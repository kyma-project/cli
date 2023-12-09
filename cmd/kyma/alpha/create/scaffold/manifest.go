package scaffold

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

func (cmd *command) generateControllerObjects() error {
	binPath, err := ensureDirExists(path.Join(cmd.opts.Directory, "/bin"))
	if err != nil {
		return err
	}

	controllerGenPath := binPath + "/controller-gen"
	controllerToolsVersion := "v0.13.0"

	if _, err := os.Stat(controllerGenPath); os.IsNotExist(err) {
		goInstallCmd := exec.Command("go", "install",
			"sigs.k8s.io/controller-tools/cmd/controller-gen@"+controllerToolsVersion)
		goInstallCmd.Env = append(os.Environ(), "GOBIN="+binPath)
		goInstallCmd.Dir = cmd.opts.Directory

		cmd.CurrentStep.LogInfo("Downloading controller-gen...")
		if err := goInstallCmd.Run(); err != nil {
			return fmt.Errorf("error downloading controller-gen: %w", err)
		}
		cmd.CurrentStep.LogInfo("Downloaded controller-gen")
	}

	controllerGenCmd := exec.Command(controllerGenPath, "rbac:roleName=manager-role", "crd", "webhook",
		"paths=./...", "output:crd:artifacts:config=config/crd/bases")
	controllerGenCmd.Dir = cmd.opts.Directory
	controllerGenCmd.Env = append(os.Environ(), "GOBIN="+binPath)

	var stdout, stderr bytes.Buffer
	controllerGenCmd.Stdout = &stdout
	controllerGenCmd.Stderr = &stderr

	if err := controllerGenCmd.Run(); err != nil {
		fmt.Printf("error running controller-gen: %s\n", err.Error())
		fmt.Printf("Stdout:\n%s\n", stdout.String())
		fmt.Printf("Stderr:\n%s\n", stderr.String())

		return fmt.Errorf("error running controller-gen: %w", err)
	}
	if _, err = os.Stat(path.Join(cmd.opts.Directory, "config", "crd", "bases")); errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("error running controller-gen: controller-gen found no crds in the directory")
	}
	return nil
}

func (cmd *command) generateManifest() error {
	kustomizationPath := path.Join(cmd.opts.Directory, "config/default")
	fSys := filesys.MakeFsOnDisk()

	cmd.CurrentStep.LogInfo("Running kustomize on config/default...")
	k := krusty.MakeKustomizer(krusty.MakeDefaultOptions())
	m, err := k.Run(fSys, kustomizationPath)
	if err != nil {
		return fmt.Errorf("error while running kustomize: %w", err)
	}
	yamlBytes, err := m.AsYaml()
	if err != nil {
		return fmt.Errorf("error while generating yaml: %w", err)
	}
	err = os.WriteFile(cmd.opts.getCompleteFilePath(fileNameManifest), yamlBytes, 0600)
	if err != nil {
		return fmt.Errorf("error while saving yaml: %w", err)
	}

	return nil
}
