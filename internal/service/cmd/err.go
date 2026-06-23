package cmd

import (
	"errors"
	"fmt"
)

var (
	ErrValidation = errors.New("validation failed")
	ErrSyntax     = fmt.Errorf("%w: syntax error", ErrValidation)
	ErrStorage    = errors.New("storage failed")
)
