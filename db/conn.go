package db

import (
	"log"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DBConfig struct {
	DSN string `yaml:"dsn"`
}

var conn *gorm.DB

func Init(cfg DBConfig) error {
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN: cfg.DSN,
	}), &gorm.Config{
		Logger: logger.New(log.New(logrus.StandardLogger().Out, "\r\n", log.LstdFlags),
			logger.Config{
				SlowThreshold:             2000 * time.Millisecond,
				LogLevel:                  logger.Warn,
				IgnoreRecordNotFoundError: false,
				Colorful:                  true,
			}),
		NowFunc: func() time.Time {
			ti, _ := time.LoadLocation("Asia/Shanghai")
			return time.Now().In(ti)
		},
	})

	if err != nil {
		logrus.Errorf("failed to connect to database: %v", err)
	} else {
		conn = db
	}

	return err
}

func GetConn() *gorm.DB {
	if conn == nil {
		logrus.Fatal("database connection is not initialized")
	}
	return conn
}
