package service

import (
	"net/http"
	"urllite/store"
	"urllite/types"
	"urllite/utils"
)

type urlService struct {
	store store.Store
}

type UrlService interface {
	CreateUrl(longUrl string) (*types.URL, *types.ApplicationError)
	GetUrlByID(id string) (*types.URL, *types.ApplicationError)
	GetUrlByShortUrl(short_url string) (*types.URL, *types.ApplicationError)
	DeleteUrlById(id string) *types.ApplicationError
	GetUrlsOfUser(user_id string) ([]*types.URL, *types.ApplicationError)
}

func NewUrlService() UrlService {
	s := store.NewStore()
	return &urlService{store: s}
}

func (u *urlService) CreateUrl(longUrl string) (*types.URL, *types.ApplicationError) {
	var url types.URL
	url.LongUrl = longUrl
	url.Status = "active"
	shortUrl, err := utils.GenerateBase62ID()
	
	if err != nil {
		return nil, &types.ApplicationError{
			Message: "Unable to generate short url",
			HttpStatusCode: http.StatusInternalServerError,
			Err: err,
		}
	}
	url.ShortUrl = shortUrl

	err = u.store.CreateURL(&url)
	if err != nil {
		return nil, &types.ApplicationError{
			Message:        "Unable to create new url",
			HttpStatusCode: http.StatusInternalServerError,
			Err:            err,
		}
	}
	return &url, nil
}

func (u *urlService) GetUrlsOfUser(user_id string) ([]*types.URL, *types.ApplicationError) {
	urls, err := u.store.GetURLsOfUser(user_id)
	if err != nil {
		return nil, &types.ApplicationError{
			Message:        "Unable to get urls",
			HttpStatusCode: http.StatusInternalServerError,
			Err:            err,
		}
	} else if len(urls) == 0 {
		return nil, &types.ApplicationError{
			Message:        "NO urls found",
			HttpStatusCode: http.StatusNotFound,
		}
	}

	return urls, nil
}

func (u *urlService) GetUrlByID(id string) (*types.URL, *types.ApplicationError) {
	url, err := u.store.GetUrlByID(id)
	if err != nil && url == nil {
		return nil, &types.ApplicationError{
			Message:        "No url found with given id",
			HttpStatusCode: http.StatusNotFound,
		}
	} else if err != nil {
		return nil, &types.ApplicationError{
			Message:        "Unable to find the url",
			HttpStatusCode: http.StatusInternalServerError,
			Err:            err,
		}
	}

	return url, nil
}

func (u *urlService) GetUrlByShortUrl(short_url string) (*types.URL, *types.ApplicationError) {
	url, err := u.store.GetUrlByShortUrl(short_url)
	if err != nil && url == nil {
		return nil, &types.ApplicationError{
			Message:        "No url found with given short url",
			HttpStatusCode: http.StatusNotFound,
		}
	} else if err != nil {
		return nil, &types.ApplicationError{
			Message:        "Unable to find the url",
			HttpStatusCode: http.StatusInternalServerError,
			Err:            err,
		}
	}

	return url, nil
}

func (u *urlService) DeleteUrlById(id string) *types.ApplicationError {
	url, appErr := u.GetUrlByID(id)
	if appErr != nil {
		return appErr
	}

	err := u.store.DeleteURL(url)
	if err != nil {
		return &types.ApplicationError{
			Message:        "Unable to delete url",
			HttpStatusCode: http.StatusInternalServerError,
			Err:            err,
		}
	}

	return nil
}
