package utils

import (
	"strings"

	"github.com/Ant-Tab-Shift/todos-service/internal/domain"
)

func Validate(task *domain.TaskSchema) error {
	if strings.TrimSpace(task.Title) == "" {
		return domain.ErrEmptyTitle
	}

	return nil
}
