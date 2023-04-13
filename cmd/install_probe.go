package cmd

import (
	"bufio"
	"bytes"
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
	Short: "Join the community powered Globalping platform by running a Docker container.",
	Long:  `Pull and run the Globalping probe Docker container on this machine. It requires Docker to be installed.`,
	Run:   installProbeCmdRun,
}

func installProbeCmdRun(cmd *cobra.Command, args []string) {
	dockerInfoCmd := exec.Command("docker", "info")
	dockerInfoCmd.Stderr = os.Stderr
	err := dockerInfoCmd.Run()
	if err != nil {
		fmt.Printf("docker info command failed: %v\n\n", err)
		fmt.Println("Docker was not detected on your system and it is required to run the Globalping probe. Please install Docker and try again.")
		return
	}

	dockerInspectCmd := exec.Command("docker", "inspect", "globalping-probe", "-f", "{{.State.Status}}")
	containerStatus, err := dockerInspectCmd.Output()
	if err == nil {
		containerStatusStr := string(bytes.TrimSpace(containerStatus))
		fmt.Printf("The globalping-probe container is already installed on your system. Current status: %s\n", containerStatusStr)
		return
	}

	ok := askUser(`The Globalping platform is a community powered project and relies on individuals like yourself to host our probes and make them accessible to everyone else.
Please confirm to pull and run our Docker container (ghcr.io/jsdelivr/globalping-probe)`)
	if !ok {
		fmt.Println("You can also run a probe manually, check our GitHub for detailed instructions. Exited without changes.")
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

	fmt.Printf("The Globalping probe started successfully. Thank you for joining our community! \n")
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
	case '\n':
		return true
	default:
		return false
	}
}
