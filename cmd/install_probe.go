package cmd

import (
	"bufio"

	"github.com/jsdelivr/globalping-cli/globalping/probe"
	"github.com/spf13/cobra"
)

func (r *Root) initInstallProbe() {
	installProbeCmd := &cobra.Command{
		Use:   "install-probe",
		Short: "Join the community-powered Globalping network by running a probe in a Docker container.",
		Long:  `The install-probe command downloads and runs the Globalping probe in a Docker container on your machine. Requires you to have Docker installed.`,
		Run:   r.RunInstallProbe,
	}

	r.Cmd.AddCommand(installProbeCmd)
}

func (r *Root) RunInstallProbe(cmd *cobra.Command, args []string) {
	containerEngine, err := r.probe.DetectContainerEngine()
	if err != nil {
		r.printer.Printf("docker info command failed: %v\n\n", err)
		r.printer.Println("Docker was not detected on your system and it is required to run the Globalping probe. Please install Docker and try again.")
		return
	}

	r.printer.Printf("Detected container engine: %s\n\n", containerEngine)

	err = r.probe.InspectContainer(containerEngine)
	if err != nil {
		r.printer.Println(err)
		return
	}

	ok := r.askUser(containerPullMessage(containerEngine))
	if !ok {
		r.printer.Println("You can also run a probe manually, check our GitHub for detailed instructions. Exited without changes.")
		return
	}

	err = r.probe.RunContainer(containerEngine)
	if err != nil {
		r.printer.Println(err)
		return
	}

	r.printer.Printf("The Globalping probe started successfully. Thank you for joining our community! \n")

	if containerEngine == probe.ContainerEnginePodman {
		r.printer.Printf("When you are using Podman, you also need to install a service to make sure the container starts on boot. Please see our instructions here: https://github.com/jsdelivr/globalping-probe/blob/master/README.md#podman-alternative\n")
	}
}

func containerPullMessage(containerEngine probe.ContainerEngine) string {
	pre := "The Globalping platform is a community powered project and relies on individuals like yourself to host our probes and make them accessible to everyone else.\n"
	var mid string
	post := "Please confirm to pull and run our Docker container (ghcr.io/jsdelivr/globalping-probe)"

	if containerEngine == probe.ContainerEnginePodman {
		mid = "We have detected that you are using podman, the 'sudo podman' command will be used to pull the container.\n"
	}

	return pre + mid + post
}

func (r *Root) askUser(s string) bool {
	r.printer.Printf("%s [Y/n] ", s)

	reader := bufio.NewReader(r.printer.InReader)

	c, _, err := reader.ReadRune()
	if err != nil {
		r.printer.Printf("failed to read character %v", err)
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
