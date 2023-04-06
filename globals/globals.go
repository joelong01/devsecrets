package globals

//
//	the is the "root" of the depdencies - so other packages can import this file, but it cannot import any packages from devsecrets
import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

//	this file builds up this structure which contains all of the settings needed to run "coral config repo" and "coral config portal"

// Globals used throughout the cli

const AAD_GROUP_NICKNAME string = "coralplatteam"
const AAD_GROUP_DESCRIPTION string = "The security group that contains Coral Platform Team"
const PLAN_NAME = "coral-portal-plan"
const ColorGreen = string("\033[32m")
const ColorReset = string("\033[0m")
const ColorRed = string("\033[31m")
const ColorYellow = string("\033[33m")
const Pass = "✓" // string ("\xE2\x9C\x94")
const Fail = "✗"

var Esc = "\x1b"

const ResetCursorPosition = "\r" //"\033[u\033[K"
const HideCursor = "\x1b[?25l"
const ShowCursor = "\x1b[?25h"
const ClearLineRight = "\x1b[0K"

var Verbose = false // verbose mode set in infra.go

type EchoLevel int

const (
	Info = iota
	Warning
	Error
)

func EchoIfVerbose(level EchoLevel, a ...any) {
	if !Verbose {
		return
	}
	switch level {
	case Info:
		EchoInfo(a...)
	case Warning:
		EchoWarning(a...)
	case Error:
		EchoError(a...)
	}

}

/*
an internal function that just echos a plain string to the console
we need this so we can filter out PATs.
*/
func internalEcho(s string) {
	fmt.Print(HidePat((s)))
}

// prints the string in red to stderr and then resets the color to normal
func EchoError(a ...any) {
	fmt.Fprint(os.Stderr, ColorRed)
	fmt.Fprint(os.Stderr, HidePat(fmt.Sprint(a...)))
	fmt.Fprint(os.Stderr, ColorReset)
}

// prints the string in yellow and then resets the color to normal
func EchoWarning(a ...any) {
	fmt.Print(ColorYellow)
	internalEcho(fmt.Sprint(a...))
	fmt.Print(ColorReset)
}

// prints the string in green and then resets the color to normal
func EchoInfo(a ...any) {
	fmt.Print(ColorGreen)
	internalEcho(fmt.Sprint(a...))
	fmt.Print(ColorReset)
}

// print in default color
func Echo(a ...any) {
	fmt.Print(ColorReset)
	internalEcho(fmt.Sprint(a...))
}

/*
prints name value pares in a nice tabular way.  Note that we do not keep track of everything printed so we can't tell
what the longest Key will be, so we just guess at 50 and would adjust it up
*/
var LongestKey int

/*
echo's to the console something that looks like

	Kind..............................................app,linux,container
	Location..........................................eastus
	port..............................................80

this is primarily used in verification
the max key width is set via running the program and picking a good number.
*/
func PrintKvp(key string, val any, valColor string) {
	// if !Verbose {
	// 	return
	// }

	const maxKeyWidth = 38
	len := len(key)
	if len > LongestKey {
		LongestKey = len
	}
	s := fmt.Sprint(key, strings.Repeat(".", maxKeyWidth-len), valColor, val, "\n", ColorReset)
	internalEcho(s)

}

func ValidationMessage(verbose bool, msg string) {
	if verbose {
		fmt.Print(ColorYellow, "Validating Input: ", msg, ColorReset, "\n")
	} else {
		fmt.Print(HideCursor, ResetCursorPosition, ClearLineRight, ColorYellow, "Validating Input: ", msg, ColorReset)
		fmt.Print(ShowCursor)
	}
}

/*
when validating a piece of data, call this first to tell the user what the app is doing.  sometimes it can take
a long time (or even crash!) and this gives the user a clue about what is going on.  In this app, the actual call
to Azure, GitHub, GitLabs, or the OS will be echo'd to stderr so this is the clue to correlate to the log line to
figure out what isn't working
*/
func PrintValidateStart(key string, valColor string) {
	const maxKeyWidth = 50
	len := len(key)
	fmt.Print(valColor, key, strings.Repeat(".", maxKeyWidth-len+1), ColorReset)
	fmt.Print(ColorReset)

}

/*
after the potentially long operation, write the resolution
*/
func PrintValidateEnd(value string, valColor string) {

	fmt.Println(valColor, value, ColorReset)

}

func Print(s string, color string) {
	fmt.Print(color + s + ColorReset)
}

// panics if the error is not nil
func PanicOnError(err error) {
	if err == nil {
		return
	}
	EchoError("Critical Error:\n" + err.Error() + "\nRerun with --verbose to get the command that may have caused the problem.")
	panic(err)
}

/*
	 prompt the user for a bool and return it
	 	True, true, T, t, Y, y all map to true
		everything else is false
*/
func EnterBoolean(prompt string, def bool) (val bool) {
	EchoWarning(prompt)
	var input string
	fmt.Scanln(&input)
	input = strings.ToLower(input)
	if input == "y" || input == "true" || input == "t" {
		val = true
	} else if input == "" {
		val = def
	} else {
		val = false
	}
	return
}

/*
prompt the user for a string and returns it
Note:  this can't use fmt.Scanln() like EnterBoolean because fmt.Scanln() splits on spaces so no words with spaces
can be entered, and it is common for the AAD Group Name to have a space in it.  if there is an error, the caller
will ignore it and just print the prompt again
*/
func EnterString(prompt string) (val string) {
	EchoWarning(prompt)
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		val = scanner.Text()
	}

	return
}

var RegExGitlabAccount = regexp.MustCompile(`glpat-[0-9a-zA-Z\\-]{20}`)
var RegExGitHubAccount = regexp.MustCompile(`^(ghp_[a-zA-Z0-9]{36}|gho_[a-zA-Z0-9]{36}|github_pat_[a-zA-Z0-9]{22}_[a-zA-Z0-9]{59}|v[0-9]\\.[0-9a-f]{40})$`)

/*
make sure that we *never* echo a PAT (github or gitlab) to the console.  the cli uses the Echo* functions
for this, so this filter will do a regex test and obfuscare the PAT if it finds one
*/
func HidePat(pat any) (val any) {
	val = pat
	switch val.(type) {
	case string:

		loc := RegExGitlabAccount.FindStringIndex(pat.(string))

		if loc != nil {

			val = fmt.Sprintf("%sglpat-********%s*****%s", pat.(string)[0:loc[0]],
				pat.(string)[loc[0]+14:loc[1]-5],
				pat.(string)[loc[1]:len(pat.(string))])
			return
		}

		loc = RegExGitHubAccount.FindStringIndex(pat.(string))

		if loc != nil {

			val = fmt.Sprintf("%sghp_**********%s**********%s", pat.(string)[0:loc[0]], // the string up to the pat
				pat.(string)[loc[0]+14:loc[1]-10],      // a subsection of the PAT, the other chars replaced by *
				pat.(string)[loc[1]:len(pat.(string))]) // the rest of the string

			return
		}

	default:
	}
	return

}
