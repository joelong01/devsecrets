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
	"log"
	"os"
	"strings"
)

/*
called by .bashrc (or .zshrc) *before* the secrets.env file is loaded.
we look at the json passed in and construct the string that should be
in the file and replace the file with the new string.

we parse the .devsecrets.env to see what the value *would be* if it was
loaded and use that instead of prompting the user.

as this runs before .bashrc or .zshrc loads the file, the only way these
environment variables are set is if something else -- e.g. codespaces --
sets the values.
*/
func OnUpdate() {
	secrets, jsonSecretFileLastModified := config.LoadSecretFile() // this also sets LocalSecrets
	envFile := config.GetSecretEnvFileName()
	fileInfo, err := os.Stat(envFile)
	scriptModTime := fileInfo.ModTime()
	globals.PanicOnError(err)
	if scriptModTime.After(jsonSecretFileLastModified) || scriptModTime.Equal(jsonSecretFileLastModified) {
		// the json hasn't changed since we last updated the secret script
		globals.EchoInfo("script is newer than json")
		// return
	}

	collectSecrets(&secrets.Secrets)

	buildAndSaveSecretsScript(secrets.Secrets)

	if config.LocalSecrets.Options.UseGitHubUserSecrets {
		saveSecretsInCodeSpace(secrets.Secrets)
	}

}

/*
iterates through the json config and get the value of the env variable
1. if they are already set in the secrets script
2. by running the configured script if it is set
3. by prompting the user
*/
func collectSecrets(secrets *[]config.Secret) {
	loadSecretsScript := config.GetSecretEnvFileName()
	var value string
	var found bool
	var err error
	for i := 0; i < len(*secrets); i++ {
		secret := &(*secrets)[i]
		found, value, _ = wrappers.FindKvpValueInFile(secret.EnvironmentVariable, loadSecretsScript)

		if !found {
			if secret.ShellScript == "" {
				prompt := fmt.Sprint("Enter value for ", secret.EnvironmentVariable, ": ")
				value = globals.EnterString(prompt)
			} else {
				value, err = wrappers.ExecBash(secret.ShellScript)
				if err != nil {
					log.Fatal("Error executing script: ", err)
				}

				if value == "" {
					globals.EchoWarning("Warning: ", secret.EnvironmentVariable,
						" was set to an empty string by the script ",
						secret.ShellScript)
				}
			}
		}

		secret.Value = value
	}
}

func buildAndSaveSecretsScript(secrets []config.Secret) {
	loadSecretsScript := config.GetSecretEnvFileName()
	toWrite := `#!/bin/bash

# if we are running in codespaces, we don't load the local environment
if [[ $CODESPACES == true ]]; then  
	return 0
fi

`
	for _, secret := range secrets {
		toWrite += fmt.Sprint("# ", secret.Description, "\n",
			secret.EnvironmentVariable, "=", "\"", secret.Value, "\"", "\n",
			"export ", secret.EnvironmentVariable, "\n")
	}

	file, err := os.Create(loadSecretsScript)
	globals.PanicOnError(err)
	defer file.Close()

	_, err = file.WriteString(toWrite)
	globals.PanicOnError(err)

}

/*
saves all the secrets in the array in CodeSpaces.  make sure to not overwrite the repos the secret is set in.
*/
func saveSecretsInCodeSpace(secrets []config.Secret) {

	pat, err := wrappers.GHGetAuthToken()
	globals.PanicOnError(err)
	// repo uses GH secrets -- make sure we are logged into GH.
	loggedIntoGithub, _ := wrappers.GetGitHubAuthStatus()
	if !loggedIntoGithub {
		err := wrappers.GHLoginToGitHub()
		if err != nil {
			globals.EchoError("Error logging into Github: ", err.Error(), "\n")
			globals.EchoError("Exiting.  Login to GH and rerun the program\n")
			os.Exit(2)
		}
	}
	ghAccountInfo, err := wrappers.GHGetAccountInfo()
	globals.PanicOnError(err)
	repos := ghAccountInfo.FullName
	for _, secret := range secrets {
		resp_code, reposForSecrets, err := wrappers.GHGetReposForSecret(secret.EnvironmentVariable, pat)
		if err != nil {
			globals.EchoError("Error processing secret ", secret.EnvironmentVariable, ".  Skipping.")
			continue
		}
		switch resp_code {
		case 200:
			for _, r := range reposForSecrets {
				if !strings.EqualFold(r.FullName, ghAccountInfo.FullName) {
					repos += "," + r.FullName
				}
			}
		case 404:
			// this will just use repos == ghAccountInfo.Repo
		default:
			globals.EchoError("Unexpected response code from calling GitHub: ", resp_code)
			continue // skip this secret
		}

		err = wrappers.GHSecretSet(ghAccountInfo.Account, ghAccountInfo.Repo, secret.EnvironmentVariable, secret.Value)
		if err != nil {
			globals.EchoError("Error saving GitHub Secrets: ", err.Error())
		}
	}

}
