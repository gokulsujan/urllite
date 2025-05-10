package service

import (
	"context"
	"net/http"
	"urllite/store"
	"urllite/types"
	"urllite/utils"

	"github.com/chromedp/chromedp"
	"github.com/gocql/gocql"
)

type urlService struct {
	store store.Store
}

type UrlService interface {
	CreateUrl(longUrl, user_id string) (*types.URL, *types.ApplicationError)
	GetUrlByID(id string) (*types.URL, *types.ApplicationError)
	GetUrlByShortUrl(short_url string) (*types.URL, *types.ApplicationError)
	DeleteUrlById(id, user_id string) *types.ApplicationError
	GetUrlsOfUser(user_id string) ([]*types.URL, *types.ApplicationError)
	GetUrlLogsByUrl(url *types.URL) ([]*types.UrlLog, *types.ApplicationError)
	GetUrlDatas(url *types.URL) (map[string]interface{}, *types.ApplicationError)
}

func NewUrlService() UrlService {
	s := store.NewStore()
	return &urlService{store: s}
}

func (u *urlService) CreateUrl(longUrl, user_id string) (*types.URL, *types.ApplicationError) {
	var url types.URL
	normalisedUrl, ok := utils.NormalizeAndValidateURL(longUrl)
	if !ok {
		return nil, &types.ApplicationError{
			Message:        "Not a valid url",
			HttpStatusCode: http.StatusBadRequest,
		}
	}

	parsedUserID, err := gocql.ParseUUID(user_id)
	if !ok {
		return nil, &types.ApplicationError{
			Message:        "Unable to find logged user data",
			HttpStatusCode: http.StatusInternalServerError,
			Err:            err,
		}
	}
	url.LongUrl = normalisedUrl
	url.Status = "active"
	url.UserID = parsedUserID
	shortUrl, err := utils.GenerateBase62ID()
	url.ShortUrl = shortUrl
	if err != nil {
		return nil, &types.ApplicationError{
			Message:        "Unable to generate short url",
			HttpStatusCode: http.StatusInternalServerError,
			Err:            err,
		}
	}

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
			Message:        "N0 urls found",
			HttpStatusCode: http.StatusNoContent,
		}
	}

	return urls, nil
}

func (u *urlService) GetUrlByID(id string) (*types.URL, *types.ApplicationError) {
	url, err := u.store.GetUrlByID(id)
	if err != nil {
		return nil, &types.ApplicationError{
			Message:        "Unable to find the user from the context",
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

func (u *urlService) DeleteUrlById(id, user_id string) *types.ApplicationError {
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

func (u *urlService) GetUrlLogsByUrl(url *types.URL) ([]*types.UrlLog, *types.ApplicationError) {
	logs, err := u.store.GetUrlLogsByUrlId(url.ID.String())
	if err != nil {
		return nil, &types.ApplicationError{
			Message:        "Unable to find logs",
			HttpStatusCode: http.StatusInternalServerError,
			Err:            err,
		}
	}

	return logs, nil
}

func (u *urlService) GetUrlDatas(url *types.URL) (map[string]interface{}, *types.ApplicationError) {
	title, favicon, err := fetchTitleAndFavicon(url.LongUrl)
	if err != nil {
		return nil, &types.ApplicationError{
			Message:        "Unable to find meta data",
			HttpStatusCode: http.StatusInternalServerError,
			Err:            err,
		}
	}

	urlInteractions, err := u.store.CountInteractions(url.ID.String())
	if err != nil {
		return nil, &types.ApplicationError{
			Message:        "Unable to get url interactions count",
			HttpStatusCode: http.StatusInternalServerError,
			Err:            err,
		}
	}
	return map[string]interface{}{"title": title, "favicon": favicon, "interactions": urlInteractions}, nil
}

func fetchTitleAndFavicon(url string) (string, string, error) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	var title, favicon string
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.Title(&title),
		chromedp.AttributeValue(`link[rel~="icon"]`, "href", &favicon, nil),
	)
	if err != nil {
		return "", "", err
	}

	return title, favicon, nil
}
