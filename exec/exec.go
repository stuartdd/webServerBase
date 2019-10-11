package exec

import (
	"errors"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/stuartdd/webServerBase/substitution"
)

/*
CmdStatus - The outcome!
*/
type CmdStatus struct {
	Stdout  string
	Stderr  string
	RetCode int
	Err     error
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
		Stdout:  "",
		Stderr:  "",
		RetCode: -1,
		Err:     nil,
	}

	zData := make([]string, len(args))
	for index, value := range args {
		zData[index] = substitution.DoSubstitution(value, data, '$')
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

	state.Stdout = strings.TrimSpace(string(sout))
	state.Stderr = strings.TrimSpace(string(serr))

	err = cmd.Wait()
	if err != nil {
		return captureError(state, err)
	}

	if state.RetCode < 0 {
		state.RetCode = 0
	}
	return state
}

func captureError(state *CmdStatus, err error) *CmdStatus {
	state.Err = err
	serr, ok := err.(*exec.ExitError)
	if ok {
		state.RetCode = serr.ExitCode()
	}
	if state.RetCode < 0 {
		state.RetCode = 1
	}
	return state
}
