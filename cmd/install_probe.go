package cmd

import (
	"bufio"
	"fmt"
	"os"

	"github.com/jsdelivr/globalping-cli/lib/probe"
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
	containerEngine, err := probe.DetectContainerEngine()
	if err != nil {
		fmt.Printf("docker info command failed: %v\n\n", err)
		fmt.Println("Docker was not detected on your system and it is required to run the Globalping probe. Please install Docker and try again.")
		return
	}

	fmt.Printf("Detected container engine: %s\n\n", containerEngine)

	err = probe.InspectContainer(containerEngine)
	if err != nil {
		fmt.Println(err)
		return
	}

	ok := askUser(`The Globalping platform is a community powered project and relies on individuals like yourself to host our probes and make them accessible to everyone else.
Please confirm to pull and run our Docker container (ghcr.io/jsdelivr/globalping-probe)`)
	if !ok {
		fmt.Println("You can also run a probe manually, check our GitHub for detailed instructions. Exited without changes.")
		return
	}

	err = probe.RunContainer(containerEngine)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("The Globalping probe started successfully. Thank you for joining our community! \n")

	if containerEngine == probe.ContainerEnginePodman {
		fmt.Printf("When you using Podman, you also need to install a service to make sure the container starts on boot. Please see our instructions here: https://github.com/jsdelivr/globalping-probe/blob/master/README.md#podman-alternative\n")
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
	case '\n':
		return true
	default:
		return false
	}
}
