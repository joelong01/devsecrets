/*
simple wrappers around "gh"
*/
package wrappers

import (
	"devsecrets/globals"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
)

/*
Note:  github names are not case sensitive!
if the repo is not found, you'll get an error that looks like:

	"Error trying to find control plane: GraphQL: Could not resolve to a Repository with the name '<repo>/coral-control-plane'. (repository)"
*/
func GHFindRepo(owner string, repoName string) (found bool, err error) {

	found = false
	args := []string{"repo", "view", fmt.Sprint(owner, "/", repoName), "--json", "name,id"}
	out, _, err := CmdExecOs("gh", args)

	if err != nil {
		if strings.Contains(err.Error(), "Could not resolve to a Repository") {
			found = false
			err = nil
		}
		return
	}

	type queryResults struct {
		Name string
		Id   string
	}
	var repo queryResults
	err = json.Unmarshal(out.Bytes(), &repo)
	if err != nil {
		return
	}

	if strings.EqualFold(repo.Name, repoName) && repo.Id != "" {
		found = true
	}
	return

}

/*
Deletes the specified repo - expand the scope if needed.  Note this is not using a PAT!
*/
func GHDeleteRepo(repoName string) (err error) {
	args := []string{"repo", "delete", repoName, "--confirm"}
	expandedScope := false
	for {
		_, _, err = CmdExecOs("gh", args)

		if err != nil {

			if !expandedScope {
				//
				//	expand the token to include delete scope
				args = []string{"auth", "refresh", "-h", "github.com", "-s", "delete_repo"}
				_, _, err = CmdExecOs("gh", args)
				if err != nil {
					err = errors.New("Error adding delete scope to gh auth token\n" + "gh error: " + err.Error())
					return
				} // -> scope expansion worked, loop and try to delete again
			} else {
				// already tried to expand scope
				err = errors.New("Could not delete repo " + repoName + " despite adding delete scope.  GH error: \n" + err.Error())
				return
			}
		} else {
			break
		}
	}
	return
}

/*
Clones the repository from a template respository without having to copy it locally

example: gh repo create coral-test-repo --template microsoft/coral-control-plane-seed --private
*/
func CloneRepositoryFromTemplate(fromRepo string, toRepo string) (err error) {
	args := []string{"repo", "create", toRepo, "--template", fromRepo, "--private"}
	_, stderr, err := CmdExecOs("gh", args)
	handleWrapperErr(&err)
	if stderr.Len() != 0 {
		err = errors.New(stderr.String())
	}
	return
}

func CreateAndPushRepo(localDirectory string, ownerOfNewRepo string, newRepoName string) (err error) {
	args := []string{"repo", "create", fmt.Sprint(ownerOfNewRepo, "/", newRepoName), "--source=" + localDirectory, "--private", "--remote=upstream", "--push"}
	_, _, err = CmdExecOs("gh", args)
	return
}

func GHCreateRepository(name string) (err error) {
	args := []string{"repo", "create", name, "--private"}
	_, _, err = CmdExecOs("gh", args)
	return
}

/*
use gh to login to GitHub.  asks for both repo and delete_repo scope.  will launch a browser to complete the
authentication.
*/
func GHLoginToGitHub() (err error) {
	args := []string{"auth", "login", "--scopes", "user,repo,codespace:secrets"}
	_, _, err = CmdExecOsInteractive("gh", args)
	if globals.Verbose && err != nil {
		globals.EchoError(err.Error() + "\n")
	}
	return
}

/*
use gh to login to GitHub, you have to read the PAT from a file, so create a temp one, write the environment
variable value to it and delete it later
*/
func GHLoginToGitHubWithPat(pat string) (err error) {
	// make a temp file and write the pat to it
	tempFile, err := os.CreateTemp("", "tmp-")
	if err != nil {
		globals.EchoError("Error creating temp file! " + err.Error())
		return
	}
	// delete temp file
	defer os.Remove(tempFile.Name())
	_, err = tempFile.Write([]byte(pat))

	// write the PAT
	if err != nil {
		globals.EchoError("Error writing PAT to temp file " + err.Error())
		return
	}

	// run a little bash script that redirects the temp file to gh auth login
	args := []string{"-c", "gh auth login --with-token < " + tempFile.Name()}
	_, _, err = CmdExecOs("bash", args)
	if globals.Verbose && err != nil {
		globals.EchoError(err.Error() + "\n")
	}

	return
}

/*
use gh to logout from GitHub.  unlike the gh command, this will not return an error if you are already logged off
*/
func LogoutFromGitHub() (err error) {
	args := []string{"auth", "logout", "--hostname", "github.com"}
	_, stderr, err := CmdExecOs("gh", args)
	if strings.Contains(stderr.String(), "not logged in") {
		err = nil
		return
	}
	if globals.Verbose && err != nil {
		globals.EchoError(err.Error() + "\n")
	}
	return
}

/*
detect if the the caller is logged into GitHub.  GitHub puts this information in stderr instead of stdout
*/
func GetGitHubAuthStatus() (loggedin bool, err error) {
	args := []string{"auth", "status"}
	_, stderr, err := CmdExecOs("gh", args)

	loggedin = false
	ghInfo := stderr.String()
	if strings.Contains(ghInfo, "Logged in to github.com as") {
		loggedin = true
		err = nil
		return
	}
	if strings.Contains(ghInfo, "You are not logged into any GitHub hosts. Run gh auth login to authenticate.") {
		loggedin = false
		err = nil
		return
	}

	if globals.Verbose && err != nil {
		globals.EchoError(err.Error() + "\n")
		return
	}

	return
}

/*
set a github secret
Coral needs this to update the GitOps repo from the control plane repo
*/
func GHSecretSet(account string, repo string, name string, secret string) (err error) {
	args := []string{"secret", "set", name, "--user", "--body", secret,
		"--repos", account + "/" + repo, "--app", "codespaces"}
	_, _, err = CmdExecOs("gh", args)
	return
}

type GithubAccountInfo struct {
	Repo    string
	Account string
}

/*
gets the repository and account names for the current project

*/

func GHGetAccountInfo() (info GithubAccountInfo, err error) {
	args := []string{"config", "--get", "remote.origin.url"}
	stdout, _, err := CmdExecOs("git", args)
	if err != nil {
		return
	}

	url := stdout.String()
	// in the format https://github.com/account/repo.git
	// Remove the ".git" extension if it exists
	if !strings.HasSuffix(url, ".git\n") {
		err = errors.New("the git config --get remote.origin.url did not return a value that ends in .git")
		return
	}
	url = url[:len(url)-5]
	// Split the URL into parts
	parts := strings.Split(url, "/")

	// Set the account and repo in the struct
	info.Account = parts[3]
	info.Repo = parts[4]

	return
}

/*
get the GitHub auth token from GH.  see
https://cli.github.com/manual/gh_auth_token
*/
func GHGetAuthToken() (token string, err error) {
	args := []string{"auth", "token"}
	stdout, _, err := CmdExecOs("gh", args)
	token = stdout.String()
	if token == "no oauth token" {
		token = ""
		err = errors.New("not logged in to GitHub")
	}

	return
}

func handleWrapperErr(err *error) (errout error) {
	if globals.Verbose && *err != nil {
		globals.EchoError((*err).Error() + "\n")
	}
	errout = *err
	return
}
