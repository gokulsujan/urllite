package tasks

import (
	"encoding/json"
	"time"

	"github.com/hibiken/asynq"
)

type urlLog struct {
}

type UrlLog interface {
	CreateLog(urlID, clientIP, country, city string, visitedAt time.Time) (*asynq.Task, error)
}

const TypeCreateUrlLog = "urllog:create"

func NewUrlLogTask() UrlLog {
	return &urlLog{}
}

func (ul *urlLog) CreateLog(urlID, clientIP, country, city string, visitedAt time.Time) (*asynq.Task, error) {
	payload, err := json.Marshal(map[string]interface{}{
		"url_id":    urlID,
		"client_ip": clientIP,
		"country":   country,
		"city":      city,
	})

	if err != nil {
		return nil, err
	}

	return asynq.NewTask(TypeCreateUrlLog, payload), nil
}
