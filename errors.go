// Copyright (c) Liam Stanley <me@liamstanley.io>. All rights reserved. Use
// of this source code is governed by the MIT license that can be found in
// the LICENSE file.

package ytdlp

func wrapError(err error) error {
	if err == nil {
		return nil
	}

	return err // TODO
}

type ErrExecution struct {
	wrapped error
}

func (e *ErrExecution) Unwrap() error {
	return e.wrapped
}

func (e *ErrExecution) Error() string {
	return "todo"
}

func IsExecutionError(err error) bool {
	_, ok := err.(*ErrExecution)
	return ok
}

type ErrParsing struct {
	wrapped error
}

func (e *ErrParsing) Unwrap() error {
	return e.wrapped
}

func (e *ErrParsing) Error() string {
	return "todo"
}

func IsParsingError(err error) bool {
	_, ok := err.(*ErrParsing)
	return ok
}

type ErrUnknown struct {
	wrapped error
}

func (e *ErrUnknown) Unwrap() error {
	return e.wrapped
}

func (e *ErrUnknown) Error() string {
	return "todo"
}

func IsUnknownError(err error) bool {
	_, ok := err.(*ErrUnknown)
	return ok
}
