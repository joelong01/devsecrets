# devsecrets
A command line utility that hooks into the VS Code container support that manages user secrets outside the project so that it reduces the chance that they secrets will get checked in.

To use this project, first create a file in the project root called devsecrets.json.  it is in this form: 
```json
{
    "options": {
        "useGitHubUserSecrets": true
    },
    "secrets": [
        {
            "environmentVariable": "GITLAB_PAT",
            "description": "The PAT for Gitlab",
            "shellscript": ""
        },
        {
            "environmentVariable": "AZURE_SUB_ID",
            "description": "An Azure subscription id used by the app",
            "shellscript": "./.devcontainer/getAzureSub.sh"
        }

    ]

}
```
Where useGitHubUserSecrets is a flag that is used to decide if all secrets will be stored in GitHub user secrets, which can then be read when using GitHub Codespaces.

The "secrets" section in the json is a simple array with 3 values that the system uses to collect the values of the secrets.

environmentVariable: the name of the env var
description: used to comment the environment variable and to prompt the user for the value of the env var
shellscript: an optional value that points to a shell script that will be executed to return the value for the env variable.  this project contains an example (getAzureSub.sh) that shows how to use it.

To integrate the system, do the following

1. copy the devsecrets binary into the container - in this example, it is in the project directory
2. update the devcontainer.json to have the following line: 
    "postCreateCommand": "./devsecrets setup --input-file devsecrets.json"
3. rebuild the container

"devsecrets setup" will update the .bashrc and the .zshrc to load the /home/vscode/devsecrets.env file and it will run "devsecrets update"

Afterwards, whenever a terminal is started devsecrets update will be called, which does the following

1. checks to see if each secret has a value
2. if not, either prompts the user for the value or executes the configured shell script to get the value
3. updates the /home/vscode/devscecrets.env file to set the environment variables for each secret.

Note that when devsecrets update runs, it will completely rewrite the .env file.  So if you want to delete a secret, remove it from the secrets array in the json and then open a new terminal. If you add a secret to the .json, the user will be prompted for its value the next time a shell is started 

Update only prompts if the value in the env var is empty - so if you want to re-prompt, edit the devsecrets.env file and delete the value (or just update it!).  Existing shells aren't updated, so you might want to close all shells after doing that.

