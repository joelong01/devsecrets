package config

import (
	"devsecrets/globals"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

/*
*

	This file contains functionality for keeping track of all the config settings. We track
		- if they've been modified so we know if we need to ask to save them or not
		- if we've already validated them so we know if we should offer to continue

	You cannot add a setting that has already been added.  You cannot remove a setting (this program doesn't need it)
	The Print() method will pretty print the settins into tabular data

*
*/

// globals
var settings = []*Setting{}

func GetAllSettings() []*Setting {
	return settings
}

// enum(s)
type Validation string

const (
	Validated Validation = "✓"
	Invalid   Validation = "✗"
	Unknown   Validation = "?"
)

type IValidateSetting interface {
	Validate() bool
}

/*
represents one setting - we convert the Cobra Flags to this simpler form.  this helps us do things
like create test settings easier.  If the value is read only (like name) or has side-effects (like
hidden) then the data is acessed by getter/setter functions, otherwise they are public. all values
are stored as strings to make things simpler
*/
type Setting struct {

	// hidden data
	validity Validation
	hidden   bool
	value    string
	name     string

	// private functions
	checkValidity func(setting *Setting) bool // to call this, call Validate()

	// public data
	Description string

	// getter functions -- needed to enable side effects
	Name   func() string
	ValueS func() string
	ValueB func() bool

	IsBool               func() bool
	Validation           func() Validation
	Hidden               func() bool
	NeedsInputValidation func() bool

	//setters - these have side effects
	SetValidated   func(Validation)
	SetDescription func(string)
	SetValue       func(any)
	SetHidden      func(bool)
}

// implement the IValidateSetting setting -- in the end this only makes it so that the dev can write
// sestting.Validate() instead of config.Validate(setting)

func (s *Setting) Validate() bool {
	if s.hidden {
		return true
	}
	if s.ValueS() == "" {
		s.validity = Invalid
		s.Description = "Setting cannot be empty string"
		return false

	}
	if !s.NeedsInputValidation() {
		return true
	}

	if s.checkValidity != nil {
		return s.checkValidity(s)
	}

	return false
}

// func New(name string, value any, description string) *Setting {
func New(name string, value string, description string, validate func(*Setting) bool) *Setting {

	s := new(Setting)
	s.name = name
	s.value = value
	s.Description = description
	s.validity = Unknown
	s.hidden = false
	s.checkValidity = validate

	s.IsBool = func() bool {
		if strings.EqualFold(s.value, "true") || strings.EqualFold(s.value, "false") {
			return true
		}
		return false
	}
	s.Name = func() string {
		return s.name
	}
	s.ValueS = func() string {
		return s.value
	}
	s.ValueB = func() bool {
		return strings.EqualFold(s.value, "true")
	}
	s.SetDescription = func(d string) {
		s.Description = d
	}

	s.Validation = func() Validation {
		return s.validity
	}

	s.SetValue = func(v any) {
		s.validity = Unknown
		s.value = fmt.Sprint(v)
		if strings.EqualFold(s.Name(), "verbose") {
			// done this way to avoid circular dependencies (config depends on global, so global can't depend on config)
			globals.Verbose = v.(bool)
		}

	}
	s.SetValidated = func(v Validation) {
		switch v {
		case Validated:
			s.validity = "✓"
		case Invalid:
			s.validity = "✗"
		case Unknown:
			s.validity = "?"
		default:
			globals.EchoError("Invalid Validation: " + fmt.Sprint(v) + "\n")
		}

	}
	s.NeedsInputValidation = func() bool {
		return !(s.validity == Validated)
	}
	s.SetHidden = func(b bool) {
		s.hidden = b
		//
		//	when something gets hidden, we move that setting to the end of the list
		//	so that the order 0...n is preserved when printing the list
		if b {
			moveToEnd(s)
		}
	}
	s.Hidden = func() bool {
		return s.hidden
	}
	if strings.EqualFold(s.Name(), "verbose") {
		// done this way to avoid circular dependencies (config depends on global, so global can't depend on config)
		globals.Verbose = s.ValueB()
	}

	return s
}

/*
this moves a setting to the end of the slice - this is important to do because we hide/show
settings for github or gitlab and it makes the table of inputs be in numerical order
*/
func moveToEnd(setting *Setting) {
	newList := []*Setting{}
	for _, s := range settings {
		if s.Name() == setting.Name() {
			continue
		}
		newList = append(newList, s)
	}
	newList = append(newList, setting)
	settings = newList
}

/*
the list is very short, so we just do a sequential search
*/
func FindSettingByName(name string) *Setting {
	for _, s := range settings {
		if s.Name() == name {
			return s
		}
	}
	return nil
}

func Value(name string) string {
	setting := FindSettingByName(name)
	if setting != nil {
		return setting.ValueS()
	}
	return ""
}

func FindSettingByIndex(index int) *Setting {
	if index >= 0 && index < len(settings) {
		return settings[index]
	}
	return nil
}

func AddSetting(name string, value string, description string) (setting *Setting, err error) {
	setting = FindSettingByName(name)

	if setting != nil {
		err = errors.New("Setting already exists, update instead of adding")
		return
	}
	setting = New(name, value, description, nil)
	settings = append(settings, setting)
	return
}

/*
cobra will gather all the various ways to set something and put it into one place
this goes through and copies that data to our internal data scructure
*/
func InitFromCobra(cmd *cobra.Command) {
	settings = nil
	settings = []*Setting{}
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		if flag.Name != "help" { // these are both flags and can't be changed at runtime
			AddSetting(flag.Name, flag.Value.String(), flag.Usage)

		}
	})
}

/*
prints a pretty table of settings and returns the number or non-hidden settings there are
*/
func PrintSettings(nameColor string, valueColor string) int {
	toPrint := make([]globals.IConsolePrint, len(settings))
	hidden := 0
	for i, setting := range settings {
		if !setting.hidden {
			toPrint[i] = setting
		} else {
			hidden++
		}
	}
	toPrint = toPrint[0 : len(toPrint)-hidden]
	header := []string{"Number", "Name", "Value", "Valid", "Description"}
	globals.PrintTable(header, toPrint)
	return len(toPrint)
}

/*
are all settings valid?
*/
func ErrorFree() bool {
	for _, s := range settings {
		if s.validity == Invalid && !s.hidden {
			return false
		}
	}
	return true
}

/*
have we modified any setting?
*/
func AnyModified() bool {
	for _, s := range settings {
		if s.validity == Unknown {
			return true
		}
	}
	return false
}

func AnyNeedLinting() bool {
	for _, s := range settings {
		if s.NeedsInputValidation() {
			return true
		}
	}
	return false
}

func Count() int {
	return len(settings)
}

func Verbose() bool {
	setting := FindSettingByName("verbose")
	if setting == nil {
		return true // this is a most likely a test
	}
	return setting.ValueB()
}

/*
the pat is in config, but as the name of the environment variable.  look it up and return the PAT
the fallback is there because the test config is different than the main config.
*/
func getPat(fallbackEnvVar string) (pat string, found bool, err error) {
	found = false
	patEnvName := FindSettingByName("gitlab-token-env-var")
	if patEnvName == nil {
		//
		//	this might be a test, in which case there are no input parameters set
		//
		pat, found = os.LookupEnv(fallbackEnvVar)
		if pat == "" || !found || err != nil {
			return "", false, errors.New("unable to configure GitHub without a valid PAT")
		} else {
			err = nil
			return
		}

	} else {
		pat, found = os.LookupEnv(patEnvName.ValueS())
	}
	return
}

func GetGitLabPat() (pat string, found bool, err error) {
	pat, found, err = getPat("GITLAB_TOKEN")
	return
}

func GetGitAccountName() (name string, found bool, err error) {
	found = false
	patEnvName := FindSettingByName("github-account-name")
	if patEnvName == nil {
		//
		//	this might be a test, in which case there are no input parameters set
		//
		name, found = os.LookupEnv("GIT_ACCOUNT_NAME") // this should be set in a '.devcontainer/local.env' file
		if name == "" || !found || err != nil {
			return "", false, errors.New("GIT_ACCOUNT_NAME environment variable not found")
		} else {
			err = nil
			return
		}

	} else {
		name, found = os.LookupEnv(patEnvName.ValueS())
	}
	return
}

func GetCreateVerifyDelete() (create bool, verify bool, delete bool) {

	switch os.Args[2] {
	case "create":
		create = true
	case "delete":
		delete = true
	case "verify":
		verify = true
	default:
		globals.EchoError("Invalid command: " + os.Args[2])
		return
	}
	return
}

// implement IConsolePrint for Setting
func (s Setting) ColumnCount() int {
	return 5
}
func (s Setting) Cell(row int, column int) string {
	switch column {
	case 0:
		return fmt.Sprint(row) // this the the "count field"
	case 1:
		return s.name
	case 2:
		return s.value
	case 3:
		return fmt.Sprint("  ", s.validity, "  ")
	case 4:
		return s.Description
	default:
		panic("Bad column index passed in")
	}
}
func (s Setting) CellColor(row int, col int) string {
	switch s.validity {
	case Validated:
		return globals.ColorGreen
	case Invalid:
		return globals.ColorRed
	case Unknown:
		return globals.ColorYellow
	default:
		panic("invalid validity")
	}
}
func (s Setting) FillChar(row int, col int) string {
	switch col {
	case 0, 1, 2:
		return "."
	case 3, 4:
		return " "
	default:
		panic("Bad column index passed in")
	}
}

type Secret struct {
	EnvironmentVariable string `json:"environmentVariable"`
	Description         string `json:"description"`
	ShellScript         string `json:"shellscript"`
}
type DevSecrets struct {
	Options struct {
		UseGitHubUserSecrets bool `json:"useGitHubUserSecrets"`
	} `json:"options"`
	Secrets []Secret `json:"secrets"`
}

var LocalSecrets DevSecrets

// implement IConsolePrint for DevSecrets
// implement IConsolePrint for Setting
func (s Secret) ColumnCount() int {
	return 3
}
func (s Secret) Cell(row int, column int) string {
	switch column {
	case 0:
		return fmt.Sprint(row) // this the the "count field"
	case 1:
		return s.EnvironmentVariable
	case 2:
		return s.Description
	case 3:
		return s.ShellScript
	default:
		panic("Bad column index passed in")
	}
}
func (s Secret) CellColor(row int, col int) string {

	return globals.ColorGreen

}
func (s Secret) FillChar(row int, col int) string {
	switch col {
	case 0, 1, 2:
		return "."
	case 3:
		return " "
	default:
		panic("Bad column index passed in")
	}
}
/*
Load the secrets file and return a string that has the file modified time
panics on error as there isn't much we can do if we can't find, open, or parse
the input file.
*/
func LoadSecretFile() (lastModified string) {
	inputFile := Value("input-file")
	if inputFile == "" {
		globals.EchoError("--input-file must be set! \n")
		os.Exit(2)
	}
	bytes, err := os.ReadFile(inputFile)
	if err != nil {
		globals.EchoError("error reading " + inputFile + " " + err.Error() + "\n")
		os.Exit(2)
	}
	err = json.Unmarshal(bytes, &LocalSecrets)
	if err != nil {
		globals.EchoError("error unmarshaling " + inputFile + " " + err.Error() + "\n")
		os.Exit(2)
	}
	fileInfo, err := os.Stat(inputFile)
	globals.PanicOnError(err)
	lastModified = fileInfo.ModTime().String()
	return
}

func GetSecretEnvFileName() (secretFileName string) {
	var secretEnvFile = ".devsecrets.sh"
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	secretFileName = filepath.Join(homeDir, secretEnvFile)
	return
}
