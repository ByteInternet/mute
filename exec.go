// Package mute implements functions to execute other programs muting std streams if required
// license: MIT, see LICENSE for details.
package mute

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
)

// ExitErrExec is exit code when failed to execute the command
const ExitErrExec = 127

// execContext is the details of an executed command
type execContext struct {
	Cmd        string
	ExitCode   int
	StdoutText *string
	StderrText *string
	Error      error
}

// Exec runs a command muting the output when matched the configuration
// executes a command, checks the exit codes and matches stdout with patterns,
// and writes the stdout/sterr when configuration did not match.
// Return the exit code of cmd, and an error if any
// bufPreAlloc is the initial size of the buffer for stdout/stderr of subcommand, in bytes
func Exec(cmd string, args []string, conf *Conf, outWriter io.Writer, errWriter io.Writer, bufPreAlloc int) (int, error) {
	if cmd == "" {
		panic("cmd is empty")
	}
	crt := cmdCriteria(cmd, conf)
	ctx := execCmd(cmd, args, bufPreAlloc)
	if !matchesCriteria(crt, ctx.ExitCode, ctx.StdoutText) {
		fmt.Fprintf(outWriter, "%v", *ctx.StdoutText)
		fmt.Fprintf(errWriter, "%v", *ctx.StderrText)
	}
	return ctx.ExitCode, ctx.Error
}

// execCmd runs the command with args and returns a pointer to an execContext
func execCmd(cmd string, args []string, bufPreAlloc int) *execContext {
	var stdoutBuffer, stderrBuffer bytes.Buffer
	if bufPreAlloc > 0 {
		stdoutBuffer.Grow(bufPreAlloc)
		stderrBuffer.Grow(bufPreAlloc)
	}
	var stdoutStr, stderrStr string
	var cmdExitCode int
	var err error
	var ctx = execContext{Cmd: cmd}
	var sigs = make(chan os.Signal, 1)

	execCmd := exec.Command(cmd, args...)
	execCmd.Stdout = &stdoutBuffer
	execCmd.Stderr = &stderrBuffer

	go func() {
		sig := <-sigs
		if execCmd.Process != nil { // signal may arrive before cmd starts
			execCmd.Process.Signal(sig)
		}
	}()
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	if err = execCmd.Run(); err != nil {
		switch e := err.(type) {
		case *exec.ExitError:
			cmdExitCode = e.ExitCode()
		default:
			cmdExitCode = ExitErrExec
		}
	}
	stdoutStr = stdoutBuffer.String()
	stdoutBuffer.Reset()
	stderrStr = stderrBuffer.String()
	stderrBuffer.Reset()
	ctx.ExitCode = cmdExitCode
	ctx.StdoutText = &stdoutStr
	ctx.StderrText = &stderrStr
	ctx.Error = err
	return &ctx
}

// matchesCriteria indicates if results of an exec matches a given Criteria
// to decide if a program should be muted or not, its exit code and stdout/stderr is matched
// against the configured Criteria. This function helps to decide on mute or not
func matchesCriteria(criteria *Criteria, code int, stdout *string) bool {
	for _, crt := range *criteria {
		if crt.IsEmpty() {
			continue
		}
		if len(crt.ExitCodes) < 1 || codesContain(crt.ExitCodes, code) {
			if len(crt.StdoutPatterns) < 1 || stdoutMatches(crt.StdoutPatterns, stdout) {
				return true
			}
		}
	}
	return false
}

// cmdCriteria returns the Criteria that the cmd should be matched against from the Conf
// Each command is matched against a criteria. The Conf has Criterias
// either per command or a default one that is used for all commands.
// cmdCriteria finds the corresponding Criterian from a Conf that the cmd
// should be checked against
func cmdCriteria(cmd string, conf *Conf) *Criteria {
	matched := ""
	for key := range conf.Commands {
		if len(key) > len(matched) && strings.HasPrefix(cmd, key) {
			matched = key
		}
	}
	if matched == "" { // no command specific criteria matched cmd
		return &conf.Default
	}
	criteria := conf.Commands[matched]
	return &criteria
}

// stdoutMatches checks if string matches any of the specified StdoutPattern regex patterns
func stdoutMatches(patterns []*StdoutPattern, stdout *string) bool {
	for _, p := range patterns {
		if p.Regexp.MatchString(*stdout) {
			return true
		}
	}
	return false
}
