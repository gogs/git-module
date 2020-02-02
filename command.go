// Copyright 2015 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package git

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"
)

// Command contains the name, arguments and environment variables of a command.
type Command struct {
	name string
	args []string
	envs []string
}

// String returns the string representation of the command.
func (c *Command) String() string {
	if len(c.args) == 0 {
		return c.name
	}
	return fmt.Sprintf("%s %s", c.name, strings.Join(c.args, " "))
}

// NewCommand creates and returns a new Command with given arguments for "git".
func NewCommand(args ...string) *Command {
	return &Command{
		name: "git",
		args: args,
	}
}

// AddArgs appends given arguments to the command.
func (c *Command) AddArgs(args ...string) *Command {
	c.args = append(c.args, args...)
	return c
}

// AddEnvs appends given environment variables to the command.
func (c *Command) AddEnvs(envs ...string) *Command {
	c.envs = append(c.envs, envs...)
	return c
}

// DefaultTimeout is the default timeout duration for all commands.
const DefaultTimeout = time.Minute

// RunInDirPipelineWithTimeout executes the command in given directory and timeout duration.
// It pipes stdout and stderr to supplied io.Writer. DefaultTimeout will be used if the timeout
// duration is less than time.Nanosecond (i.e. less than or equal to 0).
func (c *Command) RunInDirPipelineWithTimeout(timeout time.Duration, stdout, stderr io.Writer, dir string) error {
	if timeout < time.Nanosecond {
		timeout = DefaultTimeout
	}

	if len(dir) == 0 {
		log("[timeout: %v] %s", timeout, c)
	} else {
		log("[timeout: %v] %s: %s", timeout, dir, c)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, c.name, c.args...)
	if c.envs != nil {
		cmd.Env = append(os.Environ(), c.envs...)
	}
	cmd.Dir = dir
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if err := cmd.Start(); err != nil {
		return err
	}

	result := make(chan error)
	go func() {
		result <- cmd.Wait()
	}()

	select {
	case <-ctx.Done():
		if cmd.Process != nil && cmd.ProcessState != nil && !cmd.ProcessState.Exited() {
			if err := cmd.Process.Kill(); err != nil {
				return fmt.Errorf("kill process: %v", err)
			}
		}

		<-result
		return ErrExecTimeout{timeout}
	case err := <-result:
		return err
	}
}

// RunInDirWithTimeout executes the command in given directory and timeout duration.
// It returns stdout in []byte and error (combined with stderr).
func (c *Command) RunInDirWithTimeout(timeout time.Duration, dir string) ([]byte, error) {
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	if err := c.RunInDirPipelineWithTimeout(timeout, stdout, stderr, dir); err != nil {
		return nil, concatenateError(err, stderr.String())
	}

	if stdout.Len() > 0 {
		log("stdout:\n%s", stdout.Bytes()[:1024])
	}
	return stdout.Bytes(), nil
}

// RunInDirPipeline executes the command in given directory and default timeout duration.
// It pipes stdout and stderr to supplied io.Writer.
func (c *Command) RunInDirPipeline(dir string, stdout, stderr io.Writer) error {
	return c.RunInDirPipelineWithTimeout(DefaultTimeout, stdout, stderr, dir)
}

// RunInDirBytes executes the command in given directory and default timeout duration.
// It returns stdout in []byte and error (combined with stderr).
func (c *Command) RunInDirBytes(dir string) ([]byte, error) {
	return c.RunInDirWithTimeout(DefaultTimeout, dir)
}

// RunInDir executes the command in given directory and default timeout duration.
// It returns stdout in string and error (combined with stderr).
func (c *Command) RunInDir(dir string) (string, error) {
	stdout, err := c.RunInDirWithTimeout(DefaultTimeout, dir)
	if err != nil {
		return "", err
	}
	return string(stdout), nil
}

// RunWithTimeout executes the command in working directory and given timeout duration.
// It returns stdout in string and error (combined with stderr).
func (c *Command) RunWithTimeout(timeout time.Duration) (string, error) {
	stdout, err := c.RunInDirWithTimeout(timeout, "")
	if err != nil {
		return "", err
	}
	return string(stdout), nil
}

// Run executes the command in working directory and default timeout duration.
// It returns stdout in string and error (combined with stderr).
func (c *Command) Run() (string, error) {
	return c.RunWithTimeout(DefaultTimeout)
}
