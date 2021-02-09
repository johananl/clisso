package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/fatih/color"
	"github.com/howeyc/gopass"
	"github.com/mitchellh/go-homedir"

	"github.com/allcloud-io/clisso/config"
	"github.com/allcloud-io/clisso/keychain"
	"github.com/allcloud-io/clisso/platform/aws"
	"github.com/allcloud-io/clisso/provider"
	"github.com/allcloud-io/clisso/provider/okta"
	"github.com/allcloud-io/clisso/provider/onelogin"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	keyChain     = keychain.DefaultKeychain{}
	printToShell bool
	writeToFile  string
)

func init() {
	RootCmd.AddCommand(cmdGet)
	cmdGet.Flags().BoolVarP(
		&printToShell, "shell", "s", false, "Print credentials to shell",
	)
	cmdGet.Flags().StringVarP(
		&writeToFile, "write-to-file", "w", "",
		"Write credentials to this file instead of the default ($HOME/.aws/credentials)",
	)
	viper.BindPFlag("global.credentials-path", cmdGet.Flags().Lookup("write-to-file"))
}

// processCredentials prints the given Credentials to a file and/or to the shell.
func processCredentials(creds *aws.Credentials, app string) error {
	if printToShell {
		// Print credentials to shell using the correct syntax for the OS.
		aws.WriteToShell(creds, runtime.GOOS == "windows", os.Stdout)
	} else {
		path, err := homedir.Expand(viper.GetString("global.credentials-path"))
		if err != nil {
			return fmt.Errorf("expanding config file path: %v", err)
		}

		// Create the `global.credentials-path` directory if it doesn't exist.
		credsFileParentDir := filepath.Dir(path)
		if _, err := os.Stat(credsFileParentDir); os.IsNotExist(err) {
			log.Printf(color.YellowString("Credentials directory '%s' does not exist - creating it"), credsFileParentDir)

			err = os.MkdirAll(credsFileParentDir, 0755)
			if err != nil {
				return fmt.Errorf("creating credentials directory: %v", err)
			}
		}

		if err = aws.WriteToFile(creds, path, app); err != nil {
			return fmt.Errorf("writing credentials to file: %v", err)
		}
		log.Printf(color.GreenString("Credentials written successfully to '%s'"), path)
	}

	return nil
}

// sessionDuration returns a session duration using the following order of preference:
// app.duration -> provider.duration -> hardcoded default of 3600
func sessionDuration(app string) int64 {
	p := config.ProviderForApp(app)

	ad := config.AppDuration(app)
	pd := config.ProviderDuration(p)

	if ad != 0 {
		return ad
	}

	if pd != 0 {
		return pd
	}

	return 3600
}

var cmdGet = &cobra.Command{
	Use:   "get",
	Short: "Get temporary credentials for an app",
	Long: `Obtain temporary credentials for the specified app by generating a SAML
assertion at the identity provider and using this assertion to retrieve
temporary credentials from the cloud provider.

If no app is specified, the selected app (if configured) will be assumed.`,
	Run: func(cmd *cobra.Command, args []string) {
		var appName string
		if len(args) == 0 {
			// No app specified.
			selected := config.SelectedApp()
			if selected == "" {
				// No default app configured.
				log.Fatal(color.RedString("No app specified and no app selected"))
			}
			appName = selected
		} else {
			// App specified - use it.
			appName = args[0]
		}

		if pName := config.ProviderForApp(appName); pName == "" {
			log.Fatalf(color.RedString("Could not get provider for app %q"), appName)
		}

		pType := config.ProviderType(pName)
		if pType == "" {
			log.Fatalf(color.RedString("Could not get provider type for provider %q"), pName)
		}

		duration := sessionDuration(appName)

		var p provider.Provider
		var app provider.App

		switch pType {
		case provider.OneLogin:
			pc, err := onelogin.NewProviderConfig(pName)
			if err != nil {
				log.Fatalf(color.RedString("Error creating provider config: %s"), err.Error())
			}

			p, err = onelogin.New(pName, pc)
			if err != nil {
				log.Fatalf(color.RedString("Error creating provider: %s"), err.Error())
			}

			app, err = onelogin.NewApp(appName)
			if err != nil {
				log.Fatalf(color.RedString("Error creating app: %s"), err.Error())
			}
		case provider.Okta:
			pc, err := okta.NewProviderConfig(pName)
			if err != nil {
				log.Fatalf(color.RedString("Error reading provider config: %s"), err.Error())
			}

			p, err = okta.New(pName, pc)
			if err != nil {
				log.Fatalf(color.RedString("Error creating provider: %s"), err.Error())
			}

			app, err = okta.NewApp(appName)
			if err != nil {
				log.Fatalf(color.RedString("Error creating app: %s"), err.Error())
			}
		default:
			log.Fatalf(color.RedString("Unsupported identity provider type '%s' for app '%s'"), pType, app)
		}

		user := p.Username()
		if user == "" {
			fmt.Printf("%s username: ", p.Type())
			fmt.Scanln(&user)
		}
		pass, err := keyChain.Get(pName)
		if err != nil {
			fmt.Printf("%s password: ", p.Type())
			pass, err = gopass.GetPasswd()
			if err != nil {
				log.Fatalf(color.RedString("Could not read password from terminal: %v"), err)
			}
		}

		creds, err := p.Get(user, string(pass), app, duration)
		if err != nil {
			log.Fatal(color.RedString("Could not get temporary credentials: "), err)
		}
		// Process credentials
		err = processCredentials(creds, appName)
		if err != nil {
			log.Fatalf(color.RedString("Error processing credentials: %v"), err)
		}
	},
}
