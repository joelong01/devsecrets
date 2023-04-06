/*
Creates the infra necessary to run Coral.  Including
 1. the AAD App for handling AuthN
 2. the Azure resources (resource group, plan, webapp, etc) necessary to run the Coral Portal website
 3. the repos for both the Coral Portal and the GitHub control plane

the OnCreate() function will check to see if the resources arleady exist and if so, it will reuse them.  if you do not want to reust them, run
"devsecrets infra delete" prior to running "devsecrets infra create"
*/
package update

import (
	"devsecrets/config"
	"devsecrets/globals"
	"devsecrets/wrappers"
	"fmt"
	"os"
)

/*
called by .bashrc (or .zshrc) *before* the secrets.env file is loaded.
we look at the json passed in and construct the string that should be
in the file and replace the file with the new string.

if the environment variable is set, we use that value.  if not, we ask
the use what value to use
*/
func OnUpdate() {
	config.LoadSecretFile()
	var toWrite string
	for _, s := range config.LocalSecrets.Secrets {
		// is the value set?
		val := os.Getenv(s.EnvironmentVariable)
		if val == "" {
			if s.ShellScript == "" {
				prompt := fmt.Sprint("Enter value for ", s.EnvironmentVariable, ": ")
				val = globals.EnterString(prompt)
			} else {
				val, _ = wrappers.ExecBash(s.ShellScript)
			}

		}

		toWrite += fmt.Sprint("# ", s.Description, "\n",
			s.EnvironmentVariable, "=", val, "\n",
			"export ", s.EnvironmentVariable, "\n", "\n")
	}

	// globals.EchoInfo(toWrite)

	envFile := config.GetSecretFileName()
	file, err := os.Create(envFile)
	globals.PanicOnError(err)
	defer file.Close()

	_, err = file.WriteString(toWrite)
	globals.PanicOnError(err)
}
