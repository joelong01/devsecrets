/*
this files job is to
1. handle all the inputs from a file or a parameter
2. set the "Settings" struct in ../config/settings.go
3. provide a console based UI to allow the user to review/edit/verify the settings
4. provides the "usage" stdout message for the program
5. makes sure that the user is logged into Azure
*/
package setup

import (
	"devsecrets/config"
	"devsecrets/wrappers"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// SetupCmd represents the init command
var SetupCmd = &cobra.Command{
	Use:   "setup",
	Short: "bootstraps dev secrets",
	Long: ` run setup once per project to configure the system:

    Examples: 

    devsecrets setup --input-file ./devsecrets.json --verbose
    
`,
	Run: func(cmd *cobra.Command, args []string) {
		OnSetup()
	},
}

var secretEnvFile = ".devsecrets.env"

/*
arrived via 'devsecrets setup <flags>'
this should be called when the container is created.  its job is to
*. create $HOME/.devscrets.env
*. update the .bashrc to call 'devsecrets update --input-file <file> --all --verbose'
*. update the .bashrc to call 'source  $HOME/.devscrets.env'
*/
func OnSetup() error {

	// touch the .env file
	secretEnvFilePath := config.GetSecretEnvFileName()

	// update the .bashrc
	cwd, _ := os.Getwd()
	jsonSecretsInputFile := filepath.Join(cwd, config.Value("input-file"))
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	exeFileSpec, _ := os.Executable()

	secretUpdateCmd := fmt.Sprint(exeFileSpec, " update --verbose --all --input-file ", jsonSecretsInputFile)

	toWrite := fmt.Sprint(secretUpdateCmd, "\n",
		"source ", secretEnvFilePath, "\n")

	updateShellStartupFile(filepath.Join(homeDir, ".bashrc"), "devsecrets", toWrite)
	updateShellStartupFile(filepath.Join(homeDir, ".zshrc"), "devsecrets", toWrite)
	return nil
}

/*
this deletes all lines that have the word "devsecrets" in it. i'm worried that somebody or some tool might add
the executable to the path and then we blast the path.  need to fix.
*/
func updateShellStartupFile(startupFile string, deleteString string, toWrite string) {
	wrappers.RemoveLinesContaintainingString(startupFile, deleteString)
	wrappers.AppendFile(startupFile, toWrite)
}
