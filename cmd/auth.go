package cmd

import (
	"errors"
	"syscall"

	"github.com/jsdelivr/globalping-cli/globalping"
	"github.com/spf13/cobra"
)

func (r *Root) initAuth() {
	authCmd := &cobra.Command{
		Use:   "auth",
		Short: "Auth commands for the Globalping API",
	}

	loginCmd := &cobra.Command{
		RunE:  r.RunAuthLogin,
		Use:   "login",
		Short: "Authenticate with the Globalping API",
		Long:  `Authenticate with the Globalping API`,
	}

	loginFlags := loginCmd.Flags()
	loginFlags.Bool("with-token", false, "Authenticate with a token via stdin")

	statusCmd := &cobra.Command{
		RunE:  r.RunAuthStatus,
		Use:   "status",
		Short: "Check the authentication status",
		Long:  `Check the authentication status`,
	}

	logoutCmd := &cobra.Command{
		RunE:  r.RunAuthLogout,
		Use:   "logout",
		Short: "Logout from the Globalping API",
		Long:  `Logout from the Globalping API`,
	}

	authCmd.AddCommand(loginCmd)
	authCmd.AddCommand(statusCmd)
	authCmd.AddCommand(logoutCmd)

	r.Cmd.AddCommand(authCmd)
}

func (r *Root) RunAuthLogin(cmd *cobra.Command, args []string) error {
	var err error
	withToken := cmd.Flags().Changed("with-token")
	if withToken {
		return r.loginWithToken()
	}
	url := r.client.Authorize(func(e error) {
		defer func() {
			r.cancel <- syscall.SIGINT
		}()
		if e != nil {
			err = e
			return
		}
		r.printer.Println("You are now authenticated")
	})
	r.printer.Println("Please visit the following URL to authenticate:")
	r.printer.Println(url)
	<-r.cancel
	return err
}

func (r *Root) RunAuthStatus(cmd *cobra.Command, args []string) error {
	res, err := r.client.TokenIntrospection("")
	if err != nil {
		return err
	}
	if res.Active {
		r.printer.Printf("Logged in as %s.\n", res.Username)
	} else {
		r.printer.Println("Not logged in.")
	}
	return nil
}

func (r *Root) RunAuthLogout(cmd *cobra.Command, args []string) error {
	err := r.client.Logout()
	if err != nil {
		return err
	}
	r.printer.Println("You are now logged out.")
	return nil
}

func (r *Root) loginWithToken() error {
	r.printer.Println("Please enter your token:")
	token, err := r.printer.ReadPassword()
	if err != nil {
		return err
	}
	if token == "" {
		return errors.New("empty token")
	}
	introspection, err := r.client.TokenIntrospection(token)
	if err != nil {
		return err
	}
	if !introspection.Active {
		return errors.New("invalid token")
	}
	profile := r.storage.GetProfile()
	profile.Token = &globalping.Token{
		AccessToken: token,
	}
	err = r.storage.SaveConfig()
	if err != nil {
		return errors.New("failed to save token")
	}
	r.printer.Printf("Logged in as %s.\n", introspection.Username)
	return nil
}
