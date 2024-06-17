package cmd

import (
	"github.com/jsdelivr/globalping-cli/version"
	"github.com/spf13/cobra"
)

func (r *Root) initVersion() {
	r.Cmd.AddCommand(&cobra.Command{
		Run:   r.RunVersion,
		Use:   "version",
		Short: "Display the version number of your installed Globalping CLI",
	})
}

func (r *Root) RunVersion(cmd *cobra.Command, args []string) {
	r.printer.Println("Globalping CLI v" + version.Version)
}
