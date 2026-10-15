package cli

import (
	"cli-task-traker/internal/domain"
	"fmt"
	"time"
)

func FormatTask(task domain.Task) string {
	mark := " "
	if task.IsPastDeadline(time.Now()) && task.Status != "DONE" {
		mark = "!"
	}

	return fmt.Sprintf("%s %s\t%s\t%s\t%s\n",
		mark,
		task.Id,
		task.Status,
		task.Title,
		task.Deadline.Format("2006-01-02"),
	)
}
