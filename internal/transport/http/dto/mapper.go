package dto

import "github.com/Ant-Tab-Shift/todos-service/internal/domain"

func ToTaskResponse(task *domain.Task) TaskResponse {
	if task == nil {
		return TaskResponse{}
	}

	return TaskResponse{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		IsDone:      task.IsDone,
	}
}

func ToTaskListResponse(tasks []domain.Task) TaskListResponse {
	if tasks == nil {
		return TaskListResponse{
			Tasks: []TaskResponse{},
			Total: 0,
		}
	}

	responses := make([]TaskResponse, 0, len(tasks))
	for _, task := range tasks {
		responses = append(responses, ToTaskResponse(&task))
	}

	return TaskListResponse{
		Tasks: responses,
		Total: len(responses),
	}
}
