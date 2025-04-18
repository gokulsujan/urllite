package service

import (
	"net/http"
	"time"
	"urllite/store"
	"urllite/types"
)

type urlLogService struct {
	store store.Store
}

type UrlLogService interface {
	CreateUrlLogByUrl(url *types.URL, clientIp string) *types.ApplicationError
	DeleteUrlLogByUrl(urlID string) *types.ApplicationError
}

func NewUrlLogService() UrlLogService {
	s := store.NewStore()
	return &urlLogService{store: s}
}

func (uls *urlLogService) CreateUrlLogByUrl(url *types.URL, clientIp string) *types.ApplicationError {
	log := &types.UrlLog{UrlID: url.ID, VisitedAt: time.Now(), ClientIP: clientIp}
	resp, err := http.Get(url.LongUrl)
	if err != nil {
		log.HttpStatusCode = http.StatusInternalServerError
		log.RedirectStatus = err.Error()
	} else {
		log.HttpStatusCode = resp.StatusCode
		log.RedirectStatus = resp.Status
	}

	err = uls.store.CreateUrlLog(log)
	if err != nil {
		return &types.ApplicationError{
			Message:        "Unable to create the log",
			HttpStatusCode: http.StatusInternalServerError,
			Err:            err,
		}
	}

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
