package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(installProbeCmd)
}

var installProbeCmd = &cobra.Command{
	Use:   "install-probe",
	Short: "Installs and runs the Globalping probe locally in a Docker container.",
	Long: `Installs and runs the Globalping probe locally in a Docker container.
This command pulls the globalping-probe container and runs it. It requires Docker to be installed.`,
	Run: installProbeCmdRun,
}

func installProbeCmdRun(cmd *cobra.Command, args []string) {
	dockerInfoCmd := exec.Command("docker", "info")
	dockerInfoCmd.Stderr = os.Stderr
	err := dockerInfoCmd.Run()
	if err != nil {
		fmt.Printf("docker info command failed: %v\n\n", err)
		fmt.Println("Docker was not detected on your system. Docker is required to install the Globalping probe. Please install Docker and try again.")
		return
	}

	dockerInspectCmd := exec.Command("docker", "inspect", "globalping-probe", "-f", "{{.State.Status}}")
	containerStatus, err := dockerInspectCmd.Output()
	if err == nil {
		fmt.Printf("The globalping-probe container is already installed on your system. Current status: %s\n", containerStatus)
		return
	}

	ok := askUser("The globalping-probe container will now be pulled and run. Do you agree ?")
	if !ok {
		fmt.Println("globalping-probe installation not confirmed, exiting ...")
		return
	}

	dockerRunCmd := exec.Command("docker", "run", "-d", "--log-driver", "local", "--network", "host", "--restart", "always", "--name", "globalping-probe", "ghcr.io/jsdelivr/globalping-probe")
	dockerRunCmd.Stdout = os.Stdout
	dockerRunCmd.Stderr = os.Stderr
	err = dockerRunCmd.Run()
	if err != nil {
		fmt.Printf("docker info command failed: %v\n\n", err)
		return
	}
}

func askUser(s string) bool {
	fmt.Printf("%s [Y/n] ", s)

	r := bufio.NewReader(os.Stdin)

	c, _, err := r.ReadRune()
	if err != nil {
		fmt.Printf("failed to read character %v", err)
		return false
	}

	switch c {
	case 'Y':
		return true
	default:
		return false
	}
}
