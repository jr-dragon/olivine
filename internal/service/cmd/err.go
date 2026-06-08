package cmd

import "errors"

var (
	ErrValidation = errors.New("validation failed")
	ErrStorage    = errors.New("storage failed")
)
