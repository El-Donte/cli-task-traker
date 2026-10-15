package storage

import (
	"cli-task-traker/internal/domain"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"uuid"
)

type FileStorage struct {
	path string
}

func OpenFileStorage(path string) (*FileStorage, error) {
	fileStorage := &FileStorage{path: path}
	if _, err := os.Stat(fileStorage.path); os.IsNotExist(err) {
		if err := fileStorage.write(make([]domain.Task, 0)); err != nil {
			return nil, fmt.Errorf("error creating file: %v", err)
		}
	}
	return fileStorage, nil
}

func (fileStorage *FileStorage) read() ([]domain.Task, error) {
	raw, err := os.ReadFile(fileStorage.path)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}
	if len(raw) == 0 {
		return nil, nil
	}

	var tasks []domain.Task

	if err := json.Unmarshal(raw, &tasks); err != nil {
		return nil, fmt.Errorf("error unmarshalling file: %w", err)
	}

	return tasks, nil
}

func (fileStorage *FileStorage) write(tasks []domain.Task) error {
	raw, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		return fmt.Errorf("error encoding file: %w", err)
	}

	tmp := fileStorage.path + ".tmp"
	if err := os.WriteFile(tmp, raw, 0o644); err != nil {
		return fmt.Errorf("error writing to temp file: %w", err)
	}

	if err := os.Rename(tmp, fileStorage.path); err != nil {
		return fmt.Errorf("error replacing storage file: %w", err)
	}

	return nil
}

func (fileStorage *FileStorage) List() ([]domain.Task, error) {
	tasks, err := fileStorage.read()
	if err != nil {
		return nil, err
	}

	return tasks, nil
}

func (fileStorage *FileStorage) Save(task domain.Task) error {
	tasks, err := fileStorage.read()
	if err != nil {
		return err
	}

	replaced := false
	for i := range tasks {
		if tasks[i].Id == task.Id {
			tasks[i] = task

			replaced = true
			break
		}
	}
	if !replaced {
		tasks = append(tasks, task)
	}

	return fileStorage.write(tasks)
}

func (fileStorage *FileStorage) Get(id uuid.UUID) (domain.Task, error) {
	tasks, err := fileStorage.read()
	if err != nil {
		return domain.Task{}, err
	}

	for _, task := range tasks {
		if task.Id == id {
			return task, nil
		}
	}

	return domain.Task{}, domain.ErrTaskNotFound
}

func (fileStorage *FileStorage) Delete(id uuid.UUID) error {
	tasks, err := fileStorage.read()
	if err != nil {
		return err
	}

	found := false
	out := tasks[:0]
	for _, task := range tasks {
		if task.Id == id {
			found = true
			continue
		}
		out = append(out, task)
	}

	if !found {
		return domain.ErrTaskNotFound
	}

	return fileStorage.write(out)
}

func EnsureDir(path string) error {
	dir := filepath.Dir(path)
	if dir == "" || dir == "." {
		return nil
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", dir, err)
	}
	return nil
}
