// core/command.go
package core

import (
	"os/exec"
)

// Command represents an external command to be executed.
type Command struct {
	Name string
	Args []string
	// Optionally, you can add fields like Env, Dir, etc.
}

// Executor defines methods to execute commands.
type Executor interface {
	Run(cmd *Command) error
	Output(cmd *Command) ([]byte, error)
}

// RealExecutor is the default implementation of Executor using exec.Command.
type RealExecutor struct{}

func (e *RealExecutor) Run(cmd *Command) error {
	command := exec.Command(cmd.Name, cmd.Args...)
	command.Stdout = nil
	command.Stderr = nil
	return command.Run()
}

func (e *RealExecutor) Output(cmd *Command) ([]byte, error) {
	command := exec.Command(cmd.Name, cmd.Args...)
	return command.Output()
}
