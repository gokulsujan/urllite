package tasks

import (
	"fmt"
	"os"
	"time"

	"github.com/hibiken/asynq"
)

func PerformAync(task *asynq.Task) {
	client := asynq.NewClient(asynq.RedisClientOpt{Addr: os.Getenv("REDIS_ADDR")})
	defer client.Close()

	_, err := client.Enqueue(task, asynq.ProcessAt(time.Now()))
	if err != nil {
		fmt.Println("Failed to enqueue log task: %v", err)
	}
}

func PerformNow(task *asynq.Task) {
	client := asynq.NewClient(asynq.RedisClientOpt{Addr: os.Getenv("REDIS_ADDR")})
	defer client.Close()

	_, err := client.Enqueue(task)
	if err != nil {
		fmt.Println("Failed to enqueue log task: %v", err)
	}
}

func PerformAfter(task *asynq.Task, duration time.Duration) {
	client := asynq.NewClient(asynq.RedisClientOpt{Addr: os.Getenv("REDIS_ADDR")})
	defer client.Close()

	_, err := client.Enqueue(task, asynq.ProcessIn(duration))
	if err != nil {
		fmt.Println("Failed to enqueue log task: %v", err)
	}
}

func PerformLater(task *asynq.Task, time time.Time) {
	client := asynq.NewClient(asynq.RedisClientOpt{Addr: os.Getenv("REDIS_ADDR")})
	defer client.Close()

	_, err := client.Enqueue(task, asynq.ProcessAt(time))
	if err != nil {
		fmt.Println("Failed to enqueue log task: %v", err)
	}
}
