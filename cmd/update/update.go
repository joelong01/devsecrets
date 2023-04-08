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
	"bufio"
	"devsecrets/config"
	"devsecrets/globals"
	"devsecrets/wrappers"
	"fmt"
	"os"
	"strings"
)

var START_SECRET_SECTION string = "# START SECRETS"

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
	jsonSecretFileLastModified := config.LoadSecretFile()

	var ghAccountInfo wrappers.GithubAccountInfo
	var err error
	if config.LocalSecrets.Options.UseGitHubUserSecrets {
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
		ghAccountInfo, err = wrappers.GHGetAccountInfo()
		globals.PanicOnError(err)

	}

	envFile := config.GetSecretEnvFileName()
	secretsMap, err := parseSecretEnvFile(envFile)
	globals.PanicOnError(err)

	// has the input file changed since the last time we updated the secrets?
	lastUpdated, ok := secretsMap["JSON_SECRET_LAST_MODIFIED_TIME"]
	if ok {
		if strings.EqualFold(lastUpdated, jsonSecretFileLastModified) {
			globals.EchoInfo("Secrets in ", config.Value("input-file"), " have not changed.\n")
			return
		}
	}
	globals.EchoInfo("Updating Secrets based on ", config.Value("input-file"), "\n")

	// there has been an update -- we construct a new .devsecrets.sh file and replace the old one
	// update GitHub secrets as appropriate

	toWrite := `#!/bin/bash

	# if we are running in codespaces, we don't load the local environment
	if [[ $CODESPACES == true ]]; then  
		return 0
	fi
	
	`

	toWrite += START_SECRET_SECTION + "\n"

	// go through the json secrets (the "my depot requires these secrets") file
	// and get values for all of them - either from the $HOME/.devsecrets.sh file
	// or by prompting the user for the value
	for _, s := range config.LocalSecrets.Secrets {
		// is the value set?
		val, ok := secretsMap[s.EnvironmentVariable]
		if !ok {
			if s.ShellScript == "" {
				prompt := fmt.Sprint("Enter value for ", s.EnvironmentVariable, ": ")
				val = globals.EnterString(prompt)
			} else {
				val, _ = wrappers.ExecBash(s.ShellScript)
			}

		}
		toWrite += getBashEnvLines(s.EnvironmentVariable, val, s.Description)

		if config.LocalSecrets.Options.UseGitHubUserSecrets {
			err = wrappers.GHSecretSet(ghAccountInfo.Account, ghAccountInfo.Repo, s.EnvironmentVariable, val)
			if err != nil {
				globals.EchoError("Error saving GitHub Secrets: ", err.Error())
			}
		}

	}

	// the local file has one "magic" setting, which is the last modified time of the --input-file
	toWrite += getBashEnvLines("JSON_SECRET_LAST_MODIFIED_TIME", jsonSecretFileLastModified,
		"a setting used be the tool to not do work if it has already been done")

	// globals.EchoInfo(toWrite)

	file, err := os.Create(envFile)
	globals.PanicOnError(err)
	defer file.Close()

	_, err = file.WriteString(toWrite)
	globals.PanicOnError(err)

}
func getBashEnvLines(key string, val string, description string) string {
	if !strings.HasPrefix(val, "\"") {
		val = "\"" + val + "\""
	}
	return fmt.Sprint("# ", description, "\n",
		key, "=", val, "\n",
		"export ", key, "\n")
}

/*
everything before START_SECRET_SECTION is skipped
parseSecretEnvFile reads a file containing key-value pairs in the
format "key=value" and returns a map containing the parsed pairs.

	Blank lines, comments starting with "#" and lines starting with
	"export " are ignored. If a key is duplicated, an error is returned.
*/
func parseSecretEnvFile(fileName string) (secrets map[string]string, err error) {
	secrets = make(map[string]string)
	file, err := os.Open(fileName)
	if err != nil {
		if os.IsNotExist(err) {
			err = nil // this is ok

		}
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	beforeSecrets := true
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, START_SECRET_SECTION) {
			beforeSecrets = false
			continue
		}
		if beforeSecrets {
			continue
		}
		if len(line) == 0 || line[0] == '#' || strings.HasPrefix(line, "export ") {
			// Skip blank lines, comments and export statements.
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			// Invalid line format.
			return nil, fmt.Errorf("invalid line in file: %s", line)
		}
		key, value := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
		if _, ok := secrets[key]; ok {
			// Duplicate key found.
			return nil, fmt.Errorf("duplicate key in file: %s", key)
		}
		secrets[key] = strings.Trim(value, "\"")
	}
	if err := scanner.Err(); err != nil {
		// Error occurred while reading the file.
		return nil, err
	}
	return secrets, nil
}
