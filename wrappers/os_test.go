/*
these APIs are simple wrappers over the os operations that are needed for the rest of the cli
*/
package wrappers

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestCmdExecOs(t *testing.T) {
	type args struct {
		name string
		args []string
	}
	tests := []struct {
		name       string
		args       args
		wantOutput bytes.Buffer
		wantErr    bool
	}{
		{"run az", args{"az", []string{"version"}}, *bytes.NewBufferString(""), false},
		{"run bad cmd", args{"az23423424242", []string{"version"}}, *bytes.NewBufferString(""), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOutput, _, err := CmdExecOs(tt.args.name, tt.args.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("CmdExecOs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// we should get back something like this:
			// {
			// 	"azure-cli": "2.42.0",
			// 	"azure-cli-core": "2.42.0",
			// 	"azure-cli-telemetry": "1.0.8",
			// 	"extensions": {}
			// }
			if !tt.wantErr {
				j := make(map[string]any)
				err = json.Unmarshal(gotOutput.Bytes(), &j)
				if err != nil {
					t.Errorf("CmdExecOs() error = %v", err)
					return
				}
				// make sure we got the right data back...
				if j["azure-cli"] == "" || j["azure-cli-core"] == "" || j["azure-cli-telemetry"] == "" {
					t.Error("GetAzureSubscriptionList()", "unexpected output from az versio9n")
				}
			}
		})
	}
}


func TestCmdArgsToString(t *testing.T) {
	type args struct {
		cmd  string
		args []string
	}
	tests := []struct {
		name    string
		args    args
		wantOut string
	}{
		{"Find Region az command", args{"az", []string{"account", "list-locations", "--query", "[?name=='us east']"}}, "az account list-locations --query \"[?name=='us east']\""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotOut := CmdArgsToString(tt.args.cmd, tt.args.args); gotOut != tt.wantOut {
				t.Errorf("CmdArgsToString() = %v, want %v", gotOut, tt.wantOut)
			}
		})
	}
}



func TestReplaceTextInFiles(t *testing.T) {
	// create a temp file to test
	tempDir, err := os.MkdirTemp(os.TempDir(), "devsecrets-test-temp-dir")
	if err != nil {
		t.Error("Error creating temp directory! " + err.Error())
		return
	}
	tempFile, err := os.CreateTemp(tempDir, "*.yaml")
	if err != nil {
		t.Error("Error creating temp file! " + err.Error())
		return
	}
	defer os.Remove(tempDir)

	// delete temp file
	defer os.Remove(tempFile.Name())
	testRepoName := "https://github.com/test-account/test-repo"
	originalRepoName := "https://github.com/microsoft/coral-control-plane-seed/test"
	type testStruct struct {
		Name string
		Repo string
		Path string
	}
	testData := testStruct{"TestReplaceTextInFiles", originalRepoName, "templates/external-service/template.yaml"}
	toWrite, _ := yaml.Marshal(&testData)
	tempFile.Write(toWrite)

	err = ReplaceTextInFiles(tempDir, "\"*.yaml\"", originalRepoName, testRepoName)
	if err != nil {
		t.Error("Error ReplaceTextInFiles " + err.Error())
		return
	}
	out, err := os.ReadFile(tempFile.Name())
	if err != nil {
		t.Error("Error in ReadFile " + err.Error())
		return
	}
	var roundTrip testStruct
	err = yaml.Unmarshal(out, &roundTrip)
	if err != nil {
		t.Error("Error in yaml.Unmarshal " + err.Error())
		return
	}

	if roundTrip.Repo != testRepoName {
		t.Error("Error in TestReplaceTextInFiles.  Expected", testRepoName, " got ", roundTrip.Repo)
	}
}
