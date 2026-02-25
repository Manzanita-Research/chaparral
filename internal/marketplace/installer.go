package marketplace

import (
	"fmt"
	"os/exec"
)

// Install runs `claude plugin install <pluginID> --scope project` in the given repo directory.
// Returns the combined stdout+stderr output and any error.
func Install(pluginID, repoPath string) (string, error) {
	cmd := exec.Command("claude", "plugin", "install", pluginID, "--scope", "project")
	cmd.Dir = repoPath
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), fmt.Errorf("installing %s: %w\n%s", pluginID, err, out)
	}
	return string(out), nil
}
