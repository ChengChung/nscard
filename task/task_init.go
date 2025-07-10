package task

import (
	"time"

	"github.com/madflojo/tasks"
	"github.com/sirupsen/logrus"
)

type TaskConfigs []TaskConfig

type TaskConfig struct {
	TaskType string `yaml:"task_type"`

	TaskName           string `yaml:"task_name"`
	TaskIntervalMinute int    `yaml:"task_interval_min"`

	Properties map[string]interface{} `yaml:"properties"`
}

var scheduler *tasks.Scheduler

func Init(cfgs TaskConfigs) {
	if scheduler != nil {
		return
	}

	scheduler = tasks.New()

	for _, cfg := range cfgs {
		generator, ok := generators[cfg.TaskType]
		if !ok {
			continue
		}

		task := generator.New(cfg)
		if task == nil {
			continue
		}

		schTask := tasks.Task{
			Interval:          time.Duration(cfg.TaskIntervalMinute) * time.Minute,
			RunOnce:           false,
			RunSingleInstance: true,
			TaskFunc:          task.Run,
			ErrFunc: func(err error) {
				logrus.Errorf("task %s failed: %v", task.TaskName(), err)
			},
		}

		err := scheduler.AddWithID(task.TaskName(), &schTask)
		if err != nil {
			panic(err)
		}
	}
}

func Stop() {
	if scheduler != nil {
		scheduler.Stop()
		scheduler = nil
	}
}

func SchedulerNewTask() {}
