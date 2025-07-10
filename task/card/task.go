package card

import (
	"github.com/chengchung/nscard/db"
	"github.com/chengchung/nscard/task"
	"github.com/sirupsen/logrus"
)

func init() {
	task.RegisterTaskGenerator(&RenderCardTaskGenerator{})
}

type RenderCardTaskGenerator struct{}

func (g *RenderCardTaskGenerator) New(cfg task.TaskConfig) task.TaskI {
	return &RenderCardTask{
		name: cfg.TaskName,
	}
}

func (g *RenderCardTaskGenerator) TaskType() string {
	return "RenderCard"
}

type RenderCardTask struct {
	name string
}

func (t *RenderCardTask) TaskName() string {
	return t.name
}

func (t *RenderCardTask) Run() error {
	logrus.Info("rendering switch card for all users...")

	users, err := db.GetAllSwitchCardUsers()
	if err != nil {
		logrus.Errorf("failed to get switch card users: %v", err)
		return err
	}

	for _, user := range users {
		err := RenderCard(user.UserId)
		if err != nil {
			logrus.Errorf("failed to render card for user %s: %v", user.UserId, err)
			continue
		}
		logrus.Infof("Successfully rendered card for user %s", user.UserId)
	}
	return nil
}
