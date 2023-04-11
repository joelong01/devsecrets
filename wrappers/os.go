/*
these APIs are simple wrappers over the os operations that are needed for the rest of the cli
*/
package wrappers

import (
	"bufio"
	"bytes"
	"devsecrets/globals"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
)

/*
calls an arbitrary OS command

in general, if the os.Exec() failes, the error is in err
if the program that is being executed has an error, the error is in stderr
if the program echos to stdout, then the value is in stdout

this does not echoError on an error because the caller might expect an error (eg a negative test)
*/
func CmdExecOs(name string, args []string) (stdout bytes.Buffer, stderr bytes.Buffer, err error) {
	if globals.Verbose {
		// globals.Echo(CmdArgsToString(name, args) + "\n")
		s := globals.HidePat(CmdArgsToString(name, args))
		log.Println(s)
	}

	cmd := exec.Command(name, args...)

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Stdin = os.Stdin
	err = cmd.Run()

	//	we care more about the error returned by the app that is run, not the cmd.Run() error -
	//	eg. 'gh' will have an error "exit status 1" with a stderr of the actual message we care about
	if stderr.Len() != 0 && err != nil {
		// this means we executed the command just fine, it just happened to return an error
		err = errors.New(stderr.String())
	}

	return
}

func CmdExecOsInteractive(name string, args []string) (stdout bytes.Buffer, stderr bytes.Buffer, err error) {
	if globals.Verbose {
		// globals.Echo(CmdArgsToString(name, args) + "\n")
		s := globals.HidePat(CmdArgsToString(name, args))
		log.Println(s)
	}

	cmd := exec.Command(name, args...)
	//
	//	this will allow us to use the normal console for output/errors *and* capture the buffer to use in the program
	mwStdout := io.MultiWriter(os.Stdout, &stdout)
	mwStdErr := io.MultiWriter(os.Stderr, &stderr)
	cmd.Stdout = mwStdout
	cmd.Stderr = mwStdErr
	cmd.Stdin = os.Stdin
	_ = cmd.Start()
	err = cmd.Wait()

	//	we care more about the error returned by the app that is run, not the cmd.Run() error -
	//	eg. 'gh' will have an error "exit status 1" with a stderr of the actual message we care about
	if stderr.Len() != 0 && err != nil {
		// this means we executed the command just fine, it just happened to return an error
		err = errors.New(stderr.String())
	}

	return
}
func CmdExecGetJsonMap(name string, args []string) (j map[string]any, err error) {
	stdout, _, err := CmdExecOs(name, args)
	if err == nil {
		err = json.Unmarshal(stdout.Bytes(), &j)
		if globals.Verbose && err != nil {
			globals.EchoError(err.Error() + "\n")
		}
	}
	return
}

// convert the list of strings into a command line form.
// azure likes quotes around strings with spaces, so add them
// azure likes all --query parameters to have quotes - add them so user can copy and paste them to a
// terminal to rerun them
func CmdArgsToString(cmd string, args []string) (out string) {
	if cmd == "bash" {
		out = args[1]
		return
	}
	out = cmd + " "
	queryNext := false
	for _, arg := range args {
		if strings.Contains(arg, " ") || queryNext {
			arg = "\"" + arg + "\""
			queryNext = false
		}
		if cmd == "az" && (arg == "--query" || arg == "--filter") {
			queryNext = true
		}
		out += arg + " "
	}
	out = out[:len(out)-1]
	return
}

/*
find ./ -name '*.yaml' -type f -exec sed -i 's,https://github.com/microsoft/coral-control-plane-seed,https://github.com/$account/coral-control-plane,g' -- {} +
example: ReplaceTextInFiles(tempDir, "yaml", "https://github.com/microsoft/coral-control-plane-seed, https://github.com/owner/repo")

Note:  this is written as a little bash script because we only Exec one program (bash), which then execs multiple
programs and pipes the info correctly between them
*/
func ReplaceTextInFiles(startdir string, pattern string, find string, replace string) (err error) {

	if !strings.Contains(pattern, "\"") {
		// this will only be hit when somebody writes new code to call this funcdtion
		panic("ReplaceTextInFiles requires that the pattern be surrounded by quotes.  please fix, recompile, and try again")
	}

	script := fmt.Sprint("find ", startdir, " -name ", pattern, " -type f -exec sed -i 's,", find, ",", replace, ",", "g' -- {} +")
	args := []string{"-c", script}
	_, _, err = CmdExecOs("bash", args)
	return
}

func CopyDirectory(source string, destination string) (err error) {

	args := []string{"-rf", source, destination}
	stdout, stderr, err := CmdExecOs("cp", args)
	globals.EchoError(stderr.String())
	globals.EchoInfo(stdout.String())
	return
}

func Touch(file string) (err error) {
	_, _, err = CmdExecOs("touch", []string{file})
	return
}

/*
uses sed to find a string in a file.  sets found to true if found
error is set to be what sed reports to stderr.  value is in value
*/
func FindKvpValueInFile(toFind string, fileName string) (found bool, value string, err error) {
	args := []string{"-n", "-e", "s/^" + toFind + "=\"\\([^\"]*\\)\"$/\\1/p", fileName}
	stdout, stderr, err := CmdExecOs("sed", args)
	if stdout.Len() != 0 {
		found = true
		value = strings.TrimSuffix(strings.TrimPrefix(stdout.String(), "\""), "\"")
		value = strings.TrimSuffix(value, "\n")
		return
	}

	if err == nil && stderr.Len() == 0 {
		found = false
		return
	}

	if stderr.Len() != 0 {
		err = errors.New(stderr.String())
		found = false
	}
	return
}

func RemoveLinesContaintainingString(fileName string, toFind string) (err error) {
	args := []string{"-i", fmt.Sprint("/", toFind, "/d"), fileName}
	_, _, err = CmdExecOs("sed", args)
	return

}

//	func AppendFile(fileName string, text string) (err error) {
//	    args := []string{"-c", "echo", text, ">>", fileName}
//	    stdout, stderr, err := CmdExecOs("/bin/bash", args)
//	    if err != nil {
//	        return err
//	    }
//	    if stderr.Len() != 0 {
//	        err = errors.New(stderr.String())
//	        return err
//	    }
//	    if stdout.Len() > 0 {
//	        globals.EchoInfo(stdout.String(), "\n")
//	    }
//	    return nil
//	}
func AppendFile(fileName string, text string) (err error) {
	f, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.WriteString(text + "\n"); err != nil {
		return err
	}

	return nil
}

func ReplaceFile(fileName string, with string) error {
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(with)
	if err != nil {
		return err
	}

	return nil
}

/*
this executes a bash script and then returns the *last line* of the output --
so whatever the script wants to return should be the last echo call.
*/
func ExecBash(script string) (string, error) {
	// Define the command to execute the script
	var buf bytes.Buffer
	// Create a multi-writer to write to both os.Stdout and the buffer
	mw := io.MultiWriter(os.Stdout, &buf)

	cmd := &exec.Cmd{
		Path:   script,
		Stdout: mw,
		Stdin:  os.Stdin,
		Stderr: os.Stderr,
	}

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("error running script: %w", err)
	}

	// Create a scanner to read through the buffer
	scanner := bufio.NewScanner(&buf)

	// Keep scanning until the last line is reached
	var lastLine string
	for scanner.Scan() {
		lastLine = scanner.Text()
	}

	// Check for scanner errors
	err = scanner.Err()
	if err != nil {
		// Handle errors
		fmt.Println("Error reading output:", err)
		return "", err
	}

	return lastLine, nil
}
