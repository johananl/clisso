package cmd

import (
	"fmt"
	"log"
	"sort"
	"strconv"
	"syscall"

	"github.com/allcloud-io/clisso/keychain"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/term"
)

// OneLogin
var clientID string
var clientSecret string
var subdomain string
var username string
var region string
var providerDuration int

// Okta
var baseURL string
var challenges []string

func init() {
	// OneLogin
	cmdProvidersCreateOneLogin.Flags().StringVar(&clientID, "client-id", "",
		"OneLogin API client ID")
	cmdProvidersCreateOneLogin.Flags().StringVar(&clientSecret, "client-secret", "",
		"OneLogin API client secret")
	cmdProvidersCreateOneLogin.Flags().StringVar(&subdomain, "subdomain", "", "OneLogin subdomain")
	cmdProvidersCreateOneLogin.Flags().StringVar(&username, "username", "",
		"Don't ask for a username and use this instead")
	cmdProvidersCreateOneLogin.Flags().StringVar(&region, "region", "US",
		"Region in which the OneLogin API lives")
	cmdProvidersCreateOneLogin.Flags().IntVar(&providerDuration, "duration", 0, "(Optional) Default session duration in seconds")

	mandatoryFlag(cmdProvidersCreateOneLogin, "client-id")
	mandatoryFlag(cmdProvidersCreateOneLogin, "client-secret")
	mandatoryFlag(cmdProvidersCreateOneLogin, "subdomain")

	// Okta
	cmdProvidersCreateOkta.Flags().StringVar(&baseURL, "base-url", "", "Okta base URL")
	cmdProvidersCreateOkta.Flags().StringArrayVar(&challenges, "challenges", []string{"push", "token:software:totp"}, "Challenges to be used with preference")
	cmdProvidersCreateOkta.Flags().StringVar(&username, "username", "",
		"Don't ask for a username and use this instead")
	cmdProvidersCreateOkta.Flags().IntVar(&providerDuration, "duration", 0, "(Optional) Default session duration in seconds")

	mandatoryFlag(cmdProvidersCreateOkta, "base-url")

	// Build command tree
	RootCmd.AddCommand(cmdProviders)
	cmdProviders.AddCommand(cmdProvidersList)
	cmdProviders.AddCommand(cmdProvidersPassword)
	cmdProviders.AddCommand(cmdProvidersCreate)
	cmdProvidersCreate.AddCommand(cmdProvidersCreateOneLogin)
	cmdProvidersCreate.AddCommand(cmdProvidersCreateOkta)
}

var cmdProviders = &cobra.Command{
	Use:   "providers",
	Short: "Manage providers",
	Long:  `View and change provider configuration.`,
}

var cmdProvidersList = &cobra.Command{
	Use:   "ls",
	Short: "List providers",
	Long:  "List all configured providers.",
	Run: func(cmd *cobra.Command, args []string) {
		providers := viper.GetStringMap("providers")

		if len(providers) == 0 {
			log.Println("No providers configured")
			return
		}

		// Sort apps alphabetically
		keys := make([]string, 0, len(providers))
		for k := range providers {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			log.Println(k)
		}
	},
}

var cmdProvidersPassword = &cobra.Command{
	Use:   "passwd",
	Short: "Save password in KeyChain for provider",
	Long:  "Save password in KeyChain for provider, see github.com/tmc/keyring for supported stores.",
	Args:  cobra.ExactArgs(1),

	Run: func(cmd *cobra.Command, args []string) {
		provider := args[0]
		fmt.Printf("Please enter the password for the '%s' provider: ", provider)
		pass, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			log.Fatalf(color.RedString("Could not read password"))
		}

		keyChain := keychain.DefaultKeychain{}

		err = keyChain.Set(provider, pass)
		if err != nil {
			log.Fatalf("Could not save to keychain: %+v", err)
		}
		log.Printf(color.GreenString("Saved password for Provider '%s'"), provider)
	},
}

var cmdProvidersCreate = &cobra.Command{
	Use:   "create",
	Short: "Create a new provider",
	Long:  "Save a new provider into the config file.",
}

var cmdProvidersCreateOneLogin = &cobra.Command{
	Use:   "onelogin [provider name]",
	Short: "Create a new OneLogin provider",
	Long:  "Save a new OneLogin provider into the config file.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		// Verify provider doesn't exist
		if exists := viper.Get("providers." + name); exists != nil {
			log.Fatalf(color.RedString("Provider '%s' already exists"), name)
		}

		switch region {
		case "US", "EU":
		default:
			log.Fatal(color.RedString("Region must be either US or EU"))
		}

		conf := map[string]string{
			"client-id":     clientID,
			"client-secret": clientSecret,
			"subdomain":     subdomain,
			"type":          "onelogin",
			"username":      username,
			"region":        region,
		}
		if providerDuration != 0 {
			// Duration specified - validate value
			if providerDuration < 3600 || providerDuration > 43200 {
				log.Fatal(color.RedString("Invalid duration Specified. Valid values: 3600 - 43200"))
			}
			conf["duration"] = strconv.Itoa(providerDuration)
		}
		viper.Set(fmt.Sprintf("providers.%s", name), conf)

		// Write config to file
		err := viper.WriteConfig()
		if err != nil {
			log.Fatalf(color.RedString("Error writing config: %v"), err)
		}
		log.Printf(color.GreenString("Provider '%s' saved to config file"), name)
	},
}

var cmdProvidersCreateOkta = &cobra.Command{
	Use:   "okta [provider name]",
	Short: "Create a new Okta provider",
	Long:  "Save a new Okta provider into the config file.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		// Verify provider doesn't exist
		if exists := viper.Get("providers." + name); exists != nil {
			log.Fatalf(color.RedString("Provider '%s' already exists"), name)
		}

		conf := map[string]interface{}{
			"base-url":   baseURL,
			"type":       "okta",
			"username":   username,
			"challenges": challenges,
		}
		if providerDuration != 0 {
			// Duration specified - validate value
			if providerDuration < 3600 || providerDuration > 43200 {
				log.Fatal(color.RedString("Invalid duration Specified. Valid values: 3600 - 43200"))
			}
			conf["duration"] = strconv.Itoa(providerDuration)
		}
		viper.Set(fmt.Sprintf("providers.%s", name), conf)

		// Write config to file
		err := viper.WriteConfig()
		if err != nil {
			log.Fatalf(color.RedString("Error writing config: %v"), err)
		}
		log.Printf(color.GreenString("Provider '%s' saved to config file"), name)
	},
}
