package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

const (
	progname       = "genruntime_vars"

	// DISCUSS(cavcrosby): outputFileName is one example where make can also control
	// the value of the variable. Eventually, I'd like to generalize the program to
	// work with other projects too. That said, there is plenty of design work todo.
	// 
	// Initially, the thought was perhaps this program could take a struct and
	// template as arguments. Then, depending if we want to stick on using env vars
	// to gather the values for certain runtime vars, then we would have to come up
	// with a mechanism for deciding how the env var names will be determined.
	outputFileName = "runtime_vars.go"
)

const (
	ModeFile       = 0x0
	OS_READ        = 04
	OS_WRITE       = 02
	OS_EX          = 01
	OS_USER_SHIFT  = 6
	OS_GROUP_SHIFT = 3
	OS_OTH_SHIFT   = 0
	OS_USER_R      = OS_READ << OS_USER_SHIFT
	OS_USER_W      = OS_WRITE << OS_USER_SHIFT
	OS_USER_X      = OS_EX << OS_USER_SHIFT
	OS_USER_RW     = OS_USER_R | OS_USER_W
	OS_USER_RWX    = OS_USER_RW | OS_USER_X
	OS_GROUP_R     = OS_READ << OS_GROUP_SHIFT
	OS_GROUP_W     = OS_WRITE << OS_GROUP_SHIFT
	OS_GROUP_X     = OS_EX << OS_GROUP_SHIFT
	OS_GROUP_RW    = OS_GROUP_R | OS_GROUP_W
	OS_GROUP_RWX   = OS_GROUP_RW | OS_GROUP_X
	OS_OTH_R       = OS_READ << OS_OTH_SHIFT
	OS_OTH_W       = OS_WRITE << OS_OTH_SHIFT
	OS_OTH_X       = OS_EX << OS_OTH_SHIFT
	OS_OTH_RW      = OS_OTH_R | OS_OTH_W
	OS_OTH_RWX     = OS_OTH_RW | OS_OTH_X
)

var runtimeVarsTpl = `// Code generated by go generate; DO NOT EDIT.
package {{ .TargetPackage }}

var (
	progDataDir string
)

// Sets specific variables to be used at runtime.
func init() {
	{{ if .ProgDataDir -}}
		progDataDir = "{{ .ProgDataDir }}"
	{{ else -}}
		progDataDir = "/usr/local/share/debcomprt"
	{{- end }}
}
`

type runtimeVars struct {
	ProgDataDir   string
	TargetPackage string
}

// Start the main program execution.
func main() {
	args := os.Args[1:]
	for i := range args {
		arg := args[i]
		if arg != "-h" && arg != "--help" || arg == "-h" || arg == "--help" {
			fmt.Printf("Usage: %v\n\nThis program does not plan on having a functional command line interface (CLI).\n\n", progname)
		}
		if arg != "-h" && arg != "--help" {
			os.Exit(1)
		}
		os.Exit(0)
	}

	var buf bytes.Buffer
	if err := template.Must(template.New("").Parse(runtimeVarsTpl)).Execute(&buf, runtimeVars{
		ProgDataDir:   os.Getenv("PROG_DATA_DIR"),
		TargetPackage: "main",
	}); err != nil {
		log.Panic(err)
	}

	gitCmdOut, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		log.Panic(err)
	}

	if err := ioutil.WriteFile(
		filepath.Join(
			strings.TrimSuffix(string(gitCmdOut), "\n"),
			outputFileName,
		),
		buf.Bytes(),
		ModeFile|(OS_USER_R|OS_USER_W|OS_GROUP_R|OS_OTH_R),
	); err != nil {
		log.Panic(err)
	}

	os.Exit(0)
}
