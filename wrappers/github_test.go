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

