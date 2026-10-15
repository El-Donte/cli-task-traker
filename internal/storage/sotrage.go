package storage

import (
	"cli-task-traker/internal/domain"
	"uuid"
)

type Storage interface {
	List() ([]domain.Task, error)
	Save(task domain.Task) error
	Get(id uuid.UUID) (domain.Task, error)
	Delete(id uuid.UUID) error
}
