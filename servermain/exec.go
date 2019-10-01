package servermain

import (
	"io/ioutil"
	"os/exec"
	"strings"
)

/*
CmdStatus - The outcome!
*/
type CmdStatus struct {
	name    string
	args    []string
	stdout  string
	stderr  string
	retCode int
	err     error
}

/*
Run - run a command on the OS
*/
func RunAndWait(name string, args ...string) *CmdStatus {
	return run(name, args...)
}

/*
RunBackground - run a command on the OS and call back when complete
*/
func RunAndCallback(callback func(status *CmdStatus), name string, args ...string) {
	go callback(run(name, args...))
}

func run(name string, args ...string) *CmdStatus {
	state := &CmdStatus{
		name:    name,
		args:    args,
		stdout:  "",
		stderr:  "",
		retCode: -1,
		err:     nil,
	}
	cmd := exec.Command(name, args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		captureError(state, err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		captureError(state, err)
	}
	cmd.Start()
	sout, err := ioutil.ReadAll(stdout)
	if err != nil {
		captureError(state, err)
	}
	serr, err := ioutil.ReadAll(stderr)
	if err != nil {
		captureError(state, err)
	}

	state.stdout = strings.TrimSpace(string(sout))
	state.stderr = strings.TrimSpace(string(serr))

	err = cmd.Wait()
	if err != nil {
		captureError(state, err)
	}

	if state.retCode < 0 {
		state.retCode = 0
	}
	return state
}

func captureError(state *CmdStatus, err error) {
	state.err = err
	serr, ok := err.(*exec.ExitError)
	if ok {
		state.retCode = serr.ExitCode()
	}
}
