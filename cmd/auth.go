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
		Short: "Authenticate with the Globalping API",
		Long:  "Authenticate with the Globalping API for higher measurements limits.",
	}

	loginCmd := &cobra.Command{
		RunE:  r.RunAuthLogin,
		Use:   "login",
		Short: "Log in to your Globalping account",
		Long:  `Log in to your Globalping account for higher measurements limits.`,
	}

	loginFlags := loginCmd.Flags()
	loginFlags.Bool("with-token", false, "authenticate with a token read from stdin instead of the default browser-based flow")

	statusCmd := &cobra.Command{
		RunE:  r.RunAuthStatus,
		Use:   "status",
		Short: "Check the current authentication status",
		Long:  `Check the current authentication status.`,
	}

	logoutCmd := &cobra.Command{
		RunE:  r.RunAuthLogout,
		Use:   "logout",
		Short: "Log out from your Globalping account",
		Long:  `Log out from your Globalping account.`,
	}

	authCmd.AddCommand(loginCmd)
	authCmd.AddCommand(statusCmd)
	authCmd.AddCommand(logoutCmd)

	r.Cmd.AddCommand(authCmd)
}

func (r *Root) RunAuthLogin(cmd *cobra.Command, args []string) error {
	var err error
	oldToken := r.storage.GetProfile().Token
	withToken := cmd.Flags().Changed("with-token")
	if withToken {
		err := r.loginWithToken()
		if err != nil {
			return err
		}
		if oldToken != nil {
			r.client.RevokeToken(oldToken.RefreshToken)
		}
		return nil
	}
	res, err := r.client.Authorize(func(e error) {
		defer func() {
			r.cancel <- syscall.SIGINT
		}()
		if e != nil {
			err = e
			r.Cmd.SilenceUsage = true
			return
		}
		if oldToken != nil {
			r.client.RevokeToken(oldToken.RefreshToken)
		}
		r.printer.Println("Success! You are now authenticated.")
	})
	if err != nil {
		return err
	}
	r.printer.Println("Please visit the following URL to authenticate:")
	r.printer.Println(res.AuthorizeURL)
	r.utils.OpenBrowser(res.AuthorizeURL)
	r.printer.Println("\nCan't use the browser-based flow? Use \"globalping auth login --with-token\" to read a token from stdin instead.")
	<-r.cancel
	return err
}

func (r *Root) RunAuthStatus(cmd *cobra.Command, args []string) error {
	res, err := r.client.TokenIntrospection("")
	if err != nil {
		e, ok := err.(*globalping.AuthorizeError)
		if ok && e.ErrorType == "not_authorized" {
			r.printer.Println("Not logged in.")
			return nil
		}
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
