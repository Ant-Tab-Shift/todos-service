package domain

import (
	"errors"
)

var (
	ErrNotExists = errors.New("resource not found in storage")
	ErrEmptyTitle = errors.New("task must have non empty title")
)
