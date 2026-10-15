package service

import (
	"cli-task-traker/internal/domain"
	"cli-task-traker/internal/storage"
	"errors"
	"testing"
	"time"
	"uuid"
)

func newSvc(t *testing.T) *Service {
	t.Helper()
	storage.EnsureDir("tasks.json")
	store, err := storage.OpenFileStorage("tasks.json")
	if err != nil {
		t.Fatal(err)
	}
	return NewService(store)
}

func TestAddValid(t *testing.T) {
	s := newSvc(t)
	task, err := s.Add("write report", time.Date(2026, 12, 12, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	if task.Id == uuid.Nil() || task.Status != domain.Planned {
		t.Fatalf("unexpected task: %+v", task)
	}
}

func TestAddEmptyTitle(t *testing.T) {
	s := newSvc(t)
	_, err := s.Add("   ", time.Date(2026, 12, 12, 0, 0, 0, 0, time.UTC))
	if !errors.Is(err, domain.ErrEmptyTitle) {
		t.Fatalf("want ErrEmptyTitle, got %v", err)
	}
}
