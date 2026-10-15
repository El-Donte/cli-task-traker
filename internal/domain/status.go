package domain

import "slices"

type Status string

const (
	Planned    Status = "planted"
	InProgress Status = "in_progress"
	Canceled   Status = "canceled"
	Done       Status = "done"
)

var AllStatuses = []Status{Planned, InProgress, Canceled, Done}

func (s Status) IsValid() error {
	if !slices.Contains(AllStatuses, s) {
		return ErrInvalidStatus
	}
	return nil
}

func (s Status) IsFinal() bool {
	return s == Done || s == Canceled
}
