package service

import (
	"cli-task-traker/internal/domain"
	"cli-task-traker/internal/storage"
	"fmt"
	"sort"
	"strings"
	"time"
	"uuid"
)

type ListFilter struct {
	Status           *domain.Status
	OnlyPastDeadLine bool
	Search           string
}

type EditPatch struct {
	Title    *string
	Deadline *time.Time
}

type Service struct {
	store storage.Storage
}

func NewService(store storage.Storage) *Service {
	return &Service{store: store}
}

func (s *Service) Add(title string, deadline time.Time) (domain.Task, error) {
	task, err := domain.NewTask(title, deadline)
	if err != nil {
		return task, fmt.Errorf("could not create task: %w", err)
	}

	if err := s.store.Save(task); err != nil {
		return task, err
	}

	return task, nil
}

func (s *Service) Edit(id uuid.UUID, p EditPatch) (domain.Task, error) {
	task, err := s.store.Get(id)
	if err != nil {
		return task, fmt.Errorf("could not get task: %w", err)
	}

	if task.Status.IsFinal() {
		return domain.Task{}, fmt.Errorf("could not edit task: %w", err)
	}

	newTitle := ""
	newDeadline := time.Time{}

	if p.Title == nil {
		newTitle = task.Title
	} else {
		newTitle = *p.Title
	}

	if p.Deadline == nil {
		newDeadline = task.Deadline
	} else {
		newDeadline = *p.Deadline
	}

	if err := task.UpdateTask(newTitle, newDeadline, task.Status); err != nil {
		return domain.Task{}, fmt.Errorf("could not update task: %w", err)
	}

	if err := s.store.Save(task); err != nil {
		return domain.Task{}, fmt.Errorf("could not edit task: %w", err)
	}
	return task, nil
}

func (s *Service) EditStatus(id uuid.UUID, status domain.Status) (domain.Task, error) {
	task, err := s.store.Get(id)
	if err != nil {
		return task, fmt.Errorf("could not get task: %w", err)
	}

	if err := task.UpdateTask(task.Title, task.Deadline, status); err != nil {
		return domain.Task{}, fmt.Errorf("could not update task`s status: %w", err)
	}

	if err := s.store.Save(task); err != nil {
		return domain.Task{}, fmt.Errorf("could not update task`s status: %w", err)
	}

	return task, nil
}

func (s *Service) List(f ListFilter) ([]domain.Task, error) {
	list, err := s.store.List()
	if err != nil {
		return nil, fmt.Errorf("could not list tasks: %w", err)
	}

	now := time.Time{}
	search := strings.ToLower(strings.TrimSpace(f.Search))
	filtered := make([]domain.Task, 0, len(list))

	for _, task := range list {
		if f.Status != nil && *f.Status != task.Status {
			continue
		}

		if f.OnlyPastDeadLine && !task.IsPastDeadline(now) {
			continue
		}

		if search != "" && !strings.Contains(strings.ToLower(task.Title), search) {
			continue
		}

		filtered = append(filtered, task)
	}

	sort.Slice(filtered, func(i, j int) bool {
		if !filtered[i].Deadline.Equal(filtered[j].Deadline) {
			return filtered[i].Deadline.Before(filtered[j].Deadline)
		}
		return filtered[i].Id.String() < filtered[j].Id.String()
	})
	return filtered, nil
}

func (s *Service) Get(id uuid.UUID) (domain.Task, error) {
	return s.store.Get(id)
}

func (s *Service) Delete(id uuid.UUID) error {
	return s.store.Delete(id)
}

func Paginate(tasks []domain.Task, page, limit int) (pageTasks []domain.Task, total, totalPages int) {
	total = len(tasks)
	if limit <= 0 {
		limit = 20
	}

	if page <= 0 {
		page = 1
	}

	totalPages = (total + limit - 1) / limit
	if totalPages == 0 {
		totalPages = 1
	}

	start := (page - 1) * limit
	if start >= total {
		return nil, total, totalPages
	}

	end := min(start+limit, total)

	return tasks[start:end], total, totalPages
}
