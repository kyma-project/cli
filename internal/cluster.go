package internal

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func RunKubeCmd(c []string) string {
	cmd := exec.Command("kubectl", c[0:]...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error connecting to kubernetes cluster: %s\n", out)
		os.Exit(1)
	}
	return strings.Replace(string(out), "'", "", -1)
}
