package localcommand

import (
	"fmt"
	"os/exec"
	"syscall"
	"time"

	"github.com/dolfly/pty"
	"github.com/pkg/errors"
)

const (
	// DefaultCloseSignal DefaultCloseSignal
	DefaultCloseSignal = syscall.SIGINT
	// DefaultCloseTimeout DefaultCloseTimeout
	DefaultCloseTimeout = 10 * time.Second
)

// LocalCommand LocalCommand
type LocalCommand struct {
	command string
	argv    []string

	closeSignal  syscall.Signal
	closeTimeout time.Duration

	cmd       *exec.Cmd
	pty       pty.Pty
	ptyClosed chan struct{}
}

// New New
func New(command string, argv []string, options ...Option) (*LocalCommand, error) {
	cmd := exec.Command(command, argv...)

	pty, err := pty.Start(cmd)
	if err != nil {
		// todo close cmd?
		return nil, errors.Wrapf(err, "failed to start command `%s`", command)
	}
	ptyClosed := make(chan struct{})

	lcmd := &LocalCommand{
		command: command,
		argv:    argv,

		closeSignal:  DefaultCloseSignal,
		closeTimeout: DefaultCloseTimeout,

		cmd:       cmd,
		pty:       pty,
		ptyClosed: ptyClosed,
	}

	for _, option := range options {
		option(lcmd)
	}

	// When the process is closed by the user,
	// close pty so that Read() on the pty breaks with an EOF.
	go func() {
		defer func() {
			lcmd.pty.Close()
			close(lcmd.ptyClosed)
		}()

		lcmd.cmd.Wait()
	}()

	return lcmd, nil
}

func (lcmd *LocalCommand) Read(p []byte) (n int, err error) {
	return lcmd.pty.Read(p)
}

func (lcmd *LocalCommand) Write(p []byte) (n int, err error) {
	return lcmd.pty.Write(p)
}

// Close Close
func (lcmd *LocalCommand) Close() error {
	if lcmd.cmd != nil && lcmd.cmd.Process != nil {
		lcmd.cmd.Process.Signal(lcmd.closeSignal)
	}
	for {
		select {
		case <-lcmd.ptyClosed:
			return nil
		case <-lcmd.closeTimeoutC():
			lcmd.cmd.Process.Signal(syscall.SIGKILL)
		}
	}
}

// WindowTitleVariables WindowTitleVariables
func (lcmd *LocalCommand) WindowTitleVariables() map[string]interface{} {
	return map[string]interface{}{
		"command": lcmd.command,
		"argv":    lcmd.argv,
		"pid":     lcmd.cmd.Process.Pid,
	}
}

// ResizeTerminal ResizeTerminal
func (lcmd *LocalCommand) ResizeTerminal(width int, height int) error {
	rows, cols, err := pty.Getsize(lcmd.pty)
	if err != nil {
		fmt.Printf("rows(%d)cols(%d) \n", rows, cols)
		fmt.Printf("ResizeTerminal() error:%s\n", err)
	}
	pty.Setsize(lcmd.pty, &pty.Winsize{
		Rows: uint16(width),
		Cols: uint16(height),
		X:    0,
		Y:    0,
	})
	return nil
}

func (lcmd *LocalCommand) closeTimeoutC() <-chan time.Time {
	if lcmd.closeTimeout >= 0 {
		return time.After(lcmd.closeTimeout)
	}

	return make(chan time.Time)
}
