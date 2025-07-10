package task

type TaskGenerator interface {
	New(cfg TaskConfig) TaskI
	TaskType() string
}

type TaskI interface {
	TaskName() string
	Run() error
}

var generators = make(map[string]TaskGenerator)

func RegisterTaskGenerator(task TaskGenerator) {
	generators[task.TaskType()] = task
}
