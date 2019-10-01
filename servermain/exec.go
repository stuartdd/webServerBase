package servermain

import (
	"errors"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

/*
CmdStatus - The outcome!
*/
type CmdStatus struct {
	stdout  string
	stderr  string
	retCode int
	err     error
}

/*
RunAndWait - run a command on the OS
*/
func RunAndWait(path string, name string, data map[string]string, args ...string) *CmdStatus {
	return run(path, name, data, args...)
}

/*
RunAndCallback - run a command on the OS and call back when complete
*/
func RunAndCallback(callback func(status *CmdStatus), path string, name string, data map[string]string, args ...string) {
	go callback(run(path, name, data, args...))
}

func run(path string, name string, data map[string]string, args ...string) *CmdStatus {
	state := &CmdStatus{
		stdout:  "",
		stderr:  "",
		retCode: -1,
		err:     nil,
	}

	zData := make([]string, len(args))
	for index, value := range args {
		zData[index] = ReplaceDollar(value, data, '$')
	}

	cmd := exec.Command(name, zData...)
	if path != "" {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return captureError(state, errors.New("Path ["+path+"] does not exist"))
		}
		cmd.Dir = path
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return captureError(state, err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return captureError(state, err)
	}
	cmd.Start()
	sout, err := ioutil.ReadAll(stdout)
	if err != nil {
		return captureError(state, err)
	}
	serr, err := ioutil.ReadAll(stderr)
	if err != nil {
		return captureError(state, err)
	}

	state.stdout = strings.TrimSpace(string(sout))
	state.stderr = strings.TrimSpace(string(serr))

	err = cmd.Wait()
	if err != nil {
		return captureError(state, err)
	}

	if state.retCode < 0 {
		state.retCode = 0
	}
	return state
}

func captureError(state *CmdStatus, err error) *CmdStatus {
	state.err = err
	serr, ok := err.(*exec.ExitError)
	if ok {
		state.retCode = serr.ExitCode()
	}
	if state.retCode < 0 {
		state.retCode = 1
	}
	return state
}
