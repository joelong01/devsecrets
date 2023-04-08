/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"devsecrets/cmd/update"
	"devsecrets/cmd/delete"
	"devsecrets/cmd/setup"
	"devsecrets/cmd/verify"
	"devsecrets/config"
	"devsecrets/globals"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "devsecrets",
	Short: "a command line for individual dev secrets while using a shared repository",
	Long: `

	usage:

	devsecrets setup
	devsecrets update --all | --name <name> --input-file devsecrets.json --verbose
	devscecreats delete --all | --name <name>

`, PersistentPreRunE: OnPreRun,
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		//	globals.PrintKvp("Longest Key", globals.LongestKey, globals.ColorRed)
		globals.EchoInfo("")
	},
}
func OnPreRun(cmd *cobra.Command, args []string) error {
	
	// You can bind cobra and viper in a few locations, but PersistencePreRunE on the root command works well
	err := initConfig(cmd)
	if err != nil {
		globals.EchoError("Error loading config: :", err.Error(), "\n")
		os.Exit(2)
	}
	return err
}
// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

var CONFIG_FILE = "devsecrets.json" // the default name for our config
var inputFile string = ""           // the name passed in via --input-file for config
/*
called by Cobra - sets up all the parameters the cli uses
*/
func init() {

	rootCmd.AddCommand(update.UpdateCmd)
	rootCmd.AddCommand(delete.DeleteCmd)
	rootCmd.AddCommand(verify.VerifyCmd)
	rootCmd.AddCommand(setup.SetupCmd)

	// global

	rootCmd.PersistentFlags().StringVarP(&inputFile, "input-file", "i", "",
		"a json file with the settings to run the cli")
	rootCmd.PersistentFlags().BoolP("all", "a", false, "Include all items")
	rootCmd.PersistentFlags().StringP("name", "n", "", "The name of the item (optional)")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "echo actions to stderr")

	cobra.OnInitialize()
}

/*
The way that config is being used is:
- there are setting in Cobra that are set in the init() function
- there are settings that are in a config file (e.g. "local-input.yaml")
- this function uses Viper to look in various places for anything named "coral-config-settings.yaml"
- if a file is passed in via --input-file, it looks there instead
- if an environment variable is set matching the parameter's name, it picks up that value
- then it takes the data that it found and calls bindFlags, which copies the data in the Cobra structures
- finally, it passes the data to the config system which initializes itself using the Cobra structures
*/
func initConfig(cmd *cobra.Command) error {

	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return err
	}
	viper.SetConfigName(CONFIG_FILE)
	if inputFile != "" {
		viper.SetConfigFile(inputFile)
		configDir := path.Dir(inputFile)
		if configDir != "." && configDir != dir {
			viper.AddConfigPath(configDir)
		}
	}
	viper.AddConfigPath(dir)
	viper.AddConfigPath(".")
	viper.AddConfigPath("./")
	viper.AddConfigPath("$HOME")
	viper.AutomaticEnv() // read in environment variables that match
	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		globals.EchoError(err.Error() + "\n")
	}

	bindFlags(cmd)
	config.InitFromCobra(cmd)
	return nil
}

// Bind each cobra flag to its associated viper configuration (config file and environment variable)
func bindFlags(cmd *cobra.Command) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		// Determine the naming convention of the flags when represented in the config file
		configName := f.Name

		// Apply the viper config value to the flag when the flag is not set and viper has a value
		if !f.Changed && viper.IsSet(configName) {
			val := viper.Get(configName)
			v := fmt.Sprintf("%v", val)
			if f.Name == "source-control-manager" {
				if v != "github" && v != "gitlab" {
					v = "github"
				}
			}
			cmd.Flags().Set(f.Name, v)
		}
	})
}

/*
this function displays the user inputs to the console and provides a way to update any setting,
verify the setting, and allows the user to consider what they are going to do prior to doing it.
*/
func GetUserInputBeforeContinue(cmd *cobra.Command) {

	config.FindSettingByName("verbose").SetValidated(config.Validated)
	if len(os.Args) < 3 {
		globals.EchoError("usage: devsecrets init create|verify|delete <args>")
		os.Exit(5)
	}

	for {

		printSecrets()
		prompt := "[c] to continue (default).  [s] to save."
		prompt += "  Or enter a number [0.." + fmt.Sprint(len(config.LocalSecrets.Secrets)) + "] to modify a secret: "
		input := globals.EnterString(prompt)
		input = strings.ToLower(input)

		switch input {
		case "", "c": // "" makes this the default
			// it seems a bit odd to not want to lint the data under normal operation as it is far faster to find any
			// problems now than downstream  in the app.  this checks to see if there are any errors and then prompts
			// the user to make sure they are doing it on purpose.  config.ErrorFree() only checks for "Invalid", not
			// Uknown -- so if validation is never run, this will let the app just continue.
			if config.ErrorFree() {
				return
			} else {
				input = globals.EnterString("Continue with the possibly invalid data? [yNn]" + globals.ClearLineRight)
				input = strings.ToLower(input)
				if input == "y" {
					return
				}
			}

		case "s":
			// save changes back to the config file that was used
			inputFile := viper.ConfigFileUsed()
			saveSettings(inputFile)
		default:
			getUserInput(input)
		}

	}
}
func printSecrets() {
	header := []string{"Number", "Environment Varible", "Description", "Shell Script"}
	toPrint := make([]globals.IConsolePrint, len(config.LocalSecrets.Secrets))
	for i, secret := range config.LocalSecrets.Secrets {
		toPrint[i] = secret
	}
	globals.PrintTable(header, toPrint)
}

/*
user just typed in a number representing a setting.  make sure that number if valid,
print any useful information for the user, then set the value entered into the setting
returns true if the data was accepted, false if it was a bad entry
*/
func getUserInput(input string) bool {
	// user typed in a non-command string
	n, err := strconv.Atoi(input)
	if err != nil {
		return false // if it is not a number, throw it away and start the loop again
	}
	if n < 0 || n >= config.Count() {
		return false // number can't be more than the settings
	}
	setting := config.FindSettingByIndex(n)
	if setting.Hidden() {
		return false // setting can't be hidden
	}

	setting.Description = ""
	inputAccepted := printUsefulInformation(setting)
	if inputAccepted {
		return true
	}
	var val any
	prompt := "Enter value for " + setting.Name() + ": "
	if setting.IsBool() {
		val = globals.EnterBoolean(prompt, setting.ValueB())
	} else {
		val = globals.EnterString(prompt)
	}

	setting.SetValue(val)
	return true
}

/*
print out useful data that can be copied and pasted into the "Enter data" line
also if there is a shortcut to entering the data, put it here and set inputAccepted=true
if inputAccepted is false, the caller will ask the user for input
*/
func printUsefulInformation(setting *config.Setting) (inputAccepted bool) {
	inputAccepted = false

	switch setting.Name() {

	default:

	}

	return
}

/*
moves the settings into Viper system so that they can be saved by Viper to the fileName passed in so that we can test it
easier.  "--help" is a setting ownwed by the Cobra system, so that is filtered out.
*/
func saveSettings(fileName string) {
	//
	//	filter out settings that we don't want in the config file
	for i := 0; i < config.Count(); i++ {
		setting := config.FindSettingByIndex(i)
		if setting.Name() == "help" {
			continue
		}

		viper.Set(setting.Name(), setting.ValueS())
	}

	viper.SetConfigFile(fileName)
	viper.SetConfigType("yaml")
	viper.WriteConfig()

}
