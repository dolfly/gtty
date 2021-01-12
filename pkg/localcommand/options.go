package localcommand

import (
	"syscall"
	"time"
)

// Option Option
type Option func(*LocalCommand)

// WithCloseSignal WithCloseSignal
func WithCloseSignal(signal syscall.Signal) Option {
	return func(lcmd *LocalCommand) {
		lcmd.closeSignal = signal
	}
}

// WithCloseTimeout WithCloseTimeout
func WithCloseTimeout(timeout time.Duration) Option {
	return func(lcmd *LocalCommand) {
		lcmd.closeTimeout = timeout
	}
}
