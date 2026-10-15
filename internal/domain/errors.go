package domain

import "errors"

var (
	ErrEmptyTitle    = errors.New("title must not be empty")
	ErrDeadlinePast  = errors.New("deadline is in the past")
	ErrInvalidStatus = errors.New("invalid status")
	ErrTaskNotFound  = errors.New("task not found")
	ErrTaskIsFinal   = errors.New("task is done or canceled and cannot be modified")
	ErrInvalidTask   = errors.New("invalid task")
)
