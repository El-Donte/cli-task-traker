package domain

import (
	"fmt"
	"strings"
	"time"
	"uuid"
)

type Task struct {
	Id       uuid.UUID `json:"id"`
	Title    string    `json:"title"`
	Status   Status    `json:"status"`
	Deadline time.Time `json:"deadline"`
}

func NewTask(title string, deadline time.Time) (Task, error) {
	trimmedTitle := strings.TrimSpace(title)

	if trimmedTitle == "" {
		return Task{}, ErrEmptyTitle
	}

	if deadline.Before(time.Now()) {
		return Task{}, ErrDeadlinePast
	}

	return Task{
		Id:       uuid.New(),
		Title:    trimmedTitle,
		Status:   Planned,
		Deadline: deadline,
	}, nil
}

func (task *Task) IsPastDeadline(now time.Time) bool {
	return now.After(task.Deadline) && !task.Status.IsFinal()
}

func (task *Task) UpdateTask(newTitle string, newDeadline time.Time, newStatus Status) error {
	changed := false
	if newDeadline != task.Deadline {
		changed = true
	}
	if newTitle != task.Title {
		changed = true
	}
	if newStatus != task.Status {
		changed = true
	}

	if !changed {
		return fmt.Errorf("nothing to change")
	}

	trimmedTitle := strings.TrimSpace(newTitle)

	if trimmedTitle == "" {
		return ErrEmptyTitle
	}

	if newDeadline.Before(time.Now()) {
		return ErrDeadlinePast
	}

	if newStatus.IsValid() != nil {
		return ErrInvalidStatus
	}
	if task.Status.IsFinal() {
		return ErrTaskIsFinal
	}

	task.Status = newStatus
	task.Title = newTitle
	task.Deadline = newDeadline
	return nil
}

func ParseDeadline(stringTime string) (time.Time, error) {
	deadline, err := time.Parse("2006-01-02", stringTime)
	if err != nil {
		return time.Time{}, err
	}
	return deadline, nil
}
