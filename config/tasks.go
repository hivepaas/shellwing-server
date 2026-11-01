package config

import "time"

type Tasks struct {
	Queue TaskQueue `toml:"queue"`
}

type TaskQueue struct {
	Concurrency        int           `toml:"concurrency" env:"SHW_TASKS_QUEUE_CONCURRENCY" default:"10"`
	TaskCheckInterval  time.Duration `toml:"task_check_interval" env:"SHW_TASKS_QUEUE_TASK_CHECK_INTERVAL" default:"10m"`
	TaskCreateInterval time.Duration `toml:"task_create_interval" env:"SHW_TASKS_QUEUE_TASK_CREATE_INTERVAL" default:"10m"`
}
