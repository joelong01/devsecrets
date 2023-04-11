/*
simple wrappers around "gh"
*/
package wrappers

import (
	"devsecrets/globals"

	"os"
	"testing"
)

func TestGHLoginToGitHub(t *testing.T) {
	loggedIntoGithub, _ := GetGitHubAuthStatus()
	if !loggedIntoGithub {
		err := GHLoginToGitHub()
		if err != nil {
			globals.EchoError("Error logging into Github: ", err.Error(), "\n")
			globals.EchoError("Exiting.  Login to GH and rerun the program\n")
			os.Exit(2)
		}
	}
}

func TestGHGetReposForSecret(t *testing.T) {
	globals.Verbose = true
	pat, _ := GHGetAuthToken()
	code, repos, err := GHGetReposForSecret("GITLAB_TOKEN", pat)
	if err != nil {
		globals.PrintKvp("error", err.Error(), globals.ColorGreen)
	} else {
		globals.PrintKvp("code", code, globals.ColorGreen)
		for _, r := range repos {
			globals.PrintKvp("repo ", r.FullName, globals.ColorGreen)
		}
	}

	code, repos, err = GHGetReposForSecret("DOES_NOT_EXIST", pat)
	if err != nil {
		globals.PrintKvp("error", err.Error(), globals.ColorGreen)
	} else {
		globals.PrintKvp("code", code, globals.ColorGreen)
		for _, r := range repos {
			globals.PrintKvp("repo ", r.FullName, globals.ColorGreen)
		}
	}
}
