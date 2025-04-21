package service

import (
	"net/http"
	"time"
	"urllite/store"
	"urllite/tasks"
	"urllite/types"
	"urllite/utils"
)

type urlLogService struct {
	store store.Store
	task  tasks.UrlLog
}

type UrlLogService interface {
	CreateUrlLogByUrl(url *types.URL, clientIp string) *types.ApplicationError
	DeleteUrlLogByUrl(urlID string) *types.ApplicationError
}

func NewUrlLogService() UrlLogService {
	s := store.NewStore()
	t := tasks.NewUrlLogTask()
	return &urlLogService{store: s, task: t}
}

func (uls *urlLogService) CreateUrlLogByUrl(url *types.URL, clientIp string) *types.ApplicationError {
	location, err := utils.GetIPAddressLocation(clientIp)

	task, err := uls.task.CreateLog(url.ID.String(), clientIp, location["country"], location["city"], time.Now())
	if err != nil {
		return &types.ApplicationError{
			Message:        "Unable to create the log",
			HttpStatusCode: http.StatusInternalServerError,
			Err:            err,
		}
	}
	go tasks.PerformNow(task)
	return nil
}

func (uls *urlLogService) DeleteUrlLogByUrl(urlID string) *types.ApplicationError {
	url, err := uls.store.GetUrlByID(urlID)
	if err != nil {
		return &types.ApplicationError{
			Message:        "Unable to find the url",
			HttpStatusCode: http.StatusInternalServerError,
			Err:            err,
		}
	}
	err = uls.store.DeleteUrlLogsByUrlId(url.ID.String(), time.Now())

	if err != nil {
		return &types.ApplicationError{
			Message:        "Unable to delete url logs",
			HttpStatusCode: http.StatusInternalServerError,
			Err:            err,
		}
	}
	return nil
}
