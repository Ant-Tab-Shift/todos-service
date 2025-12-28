package domain

type TaskSchema struct {
	Title       string
	Description string
	IsDone      bool
}

type Task struct {
	ID    uint64
	TaskSchema
}
