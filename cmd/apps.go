package cmd

import (
	"fmt"
	"log"
	"sort"
	"strconv"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// TODO Get rid of these global variables.
// Common
var pName string
var duration int

// OneLogin
var appID string

// Okta
var URL string

func init() {
	// OneLogin
	cmdAppsCreateOneLogin.Flags().StringVar(&appID, "app-id", "", "OneLogin app ID")
	cmdAppsCreateOneLogin.Flags().StringVar(&pName, "provider", "", "Name of the Clisso provider")
	cmdAppsCreateOneLogin.Flags().IntVar(&duration, "duration", 0, "(Optional) Session duration in seconds")
	cmdAppsCreateOneLogin.MarkFlagRequired("app-id")
	cmdAppsCreateOneLogin.MarkFlagRequired("provider")

	// Okta
	cmdAppsCreateOkta.Flags().StringVar(&pName, "provider", "", "Name of the Clisso provider")
	cmdAppsCreateOkta.Flags().StringVar(&URL, "url", "", "Okta app URL")
	cmdAppsCreateOkta.Flags().IntVar(&duration, "duration", 0, "(Optional) Session duration in seconds")
	cmdAppsCreateOkta.MarkFlagRequired("provider")
	cmdAppsCreateOkta.MarkFlagRequired("url")

	// Build command tree
	RootCmd.AddCommand(cmdApps)
	cmdApps.AddCommand(cmdAppsList)
	cmdApps.AddCommand(cmdAppsCreate)
	cmdAppsCreate.AddCommand(cmdAppsCreateOneLogin)
	cmdAppsCreate.AddCommand(cmdAppsCreateOkta)
	cmdApps.AddCommand(cmdAppsSelect)
}

var cmdApps = &cobra.Command{
	Use:   "apps",
	Short: "Manage apps",
	Long:  `View and change app configuration.`,
}

var cmdAppsList = &cobra.Command{
	Use:   "ls",
	Short: "List apps",
	Long:  "List all configured apps.",
	Run: func(cmd *cobra.Command, args []string) {
		apps := viper.GetStringMap("apps")

		if len(apps) == 0 {
			fmt.Println("No apps configured")
			return
		}

		// Sort apps alphabetically
		keys := make([]string, 0, len(apps))
		for k := range apps {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		selected := viper.GetString("global.selected-app")

		for _, k := range keys {
			if k == selected {
				log.Printf(color.GreenString("* %s"), k)
			} else {
				log.Printf("  %s", k)
			}
		}
	},
}

var cmdAppsCreate = &cobra.Command{
	Use:   "create",
	Short: "Create a new app",
	Long:  "Save a new app into the config file.",
}

var cmdAppsCreateOneLogin = &cobra.Command{
	Use:   "onelogin [app name]",
	Short: "Create a new OneLogin app",
	Long:  "Save a new OneLogin app into the config file.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		// Verify app doesn't exist
		if exists := viper.Get("apps." + name); exists != nil {
			log.Fatalf(color.RedString("App '%s' already exists"), name)
		}

		// Verify provider exists
		if exists := viper.Get("providers." + pName); exists == nil {
			log.Fatalf(color.RedString("Provider '%s' doesn't exist"), pName)
		}

		// Verify provider type
		pType := viper.GetString(fmt.Sprintf("providers.%s.type", pName))
		if pType != "onelogin" {
			log.Fatalf(
				color.RedString("Invalid provider type '%s' for a OneLogin app. Type must be 'onelogin'."),
				pType,
			)
		}

		conf := map[string]string{
			"app-id":   appID,
			"provider": pName,
		}

		if duration != 0 {
			// Duration specified - validate value
			if duration < 3600 || duration > 43200 {
				log.Fatal(color.RedString("Invalid duration Specified. Valid values: 3600 - 43200"))
			}
			conf["duration"] = strconv.Itoa(duration)
		}

		viper.Set(fmt.Sprintf("apps.%s", name), conf)

		// Write config to file
		err := viper.WriteConfig()
		if err != nil {
			log.Fatalf(color.RedString("Error writing config: %v"), err)
		}
		log.Printf(color.GreenString("App '%s' saved to config file"), name)
	},
}

var cmdAppsCreateOkta = &cobra.Command{
	Use:   "okta [app name]",
	Short: "Create a new Okta app",
	Long:  "Save a new Okta app into the config file.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		// Verify app doesn't exist
		if exists := viper.Get("apps." + name); exists != nil {
			log.Fatalf(color.RedString("App '%s' already exists"), name)
		}

		// Verify provider exists
		if exists := viper.Get("providers." + pName); exists == nil {
			log.Fatalf(color.RedString("Provider '%s' doesn't exist"), pName)
		}

		// Verify provider type
		pType := viper.GetString(fmt.Sprintf("providers.%s.type", pName))
		if pType != "okta" {
			log.Fatalf(
				color.RedString("Invalid provider type '%s' for an Okta app. Type must be 'okta'."),
				pType,
			)
		}

		conf := map[string]string{
			"provider": pName,
			"url":      URL,
		}

		if duration != 0 {
			// Duration specified - validate value
			if duration < 3600 || duration > 43200 {
				log.Fatal(color.RedString("Invalid duration Specified. Valid values: 3600 - 43200"))
			}
			conf["duration"] = strconv.Itoa(duration)
		}

		viper.Set(fmt.Sprintf("apps.%s", name), conf)

		// Write config to file
		err := viper.WriteConfig()
		if err != nil {
			log.Fatalf(color.RedString("Error writing config: %v"), err)
		}
		log.Printf(color.GreenString("App '%s' saved to config file"), name)
	},
}

var cmdAppsSelect = &cobra.Command{
	Use:   "select [app name]",
	Short: "Select an app to be used by default",
	Long:  "Use the specified app when running `clisso get` without providing an app.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		app := args[0]

		if app == "" {
			viper.Set("global.selected-app", "")
			log.Println(color.GreenString("Unsetting selected app"))
		} else {
			if exists := viper.Get("apps." + app); exists == nil {
				log.Fatalf(color.RedString("App '%s' doesn't exist"), app)
			}
			log.Printf(color.GreenString("Setting selected app to '%s'"), app)
			viper.Set("global.selected-app", app)
		}

		// Write config to file
		err := viper.WriteConfig()
		if err != nil {
			log.Fatalf(color.RedString("Error writing config: %v"), err)
		}
	},
}
