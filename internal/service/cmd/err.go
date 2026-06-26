package cmd

import (
	"errors"
	"fmt"
)

var (
	ErrValidation = errors.New("validation failed")
	ErrSyntax     = fmt.Errorf("%w: syntax error", ErrValidation)
	ErrWrongType  = fmt.Errorf("%w: Operation against a key holding the wrong kind of value", ErrValidation)

	ErrStorage = errors.New("storage failed")
)
