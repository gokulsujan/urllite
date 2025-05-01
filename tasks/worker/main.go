package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
	"urllite/config/env"
	"urllite/store"
	"urllite/tasks"
	"urllite/types"

	"github.com/gocql/gocql"
	"github.com/hibiken/asynq"
)

func main() {
	env.EnableEnvVariables()
	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: os.Getenv("REDIS_ADDR")},
		asynq.Config{Concurrency: 10},
	)
	mux := asynq.NewServeMux()
	mux.HandleFunc(tasks.TypeCreateUrlLog, func(ctx context.Context, task *asynq.Task) error {
		var p map[string]interface{}
		if err := json.Unmarshal(task.Payload(), &p); err != nil {
			return err
		}

		urlId, ok := p["url_id"].(string)
		if !ok {
			fmt.Println("Something went wrong for url id")
		}
		clientIP, ok := p["client_ip"].(string)
		if !ok {
			fmt.Println("Something went wrong for client ip: " + clientIP)
		}
		city, ok := p["city"].(string)
		if !ok {
			fmt.Println("Something went wrong for city: " + city)
		}
		country, ok := p["country"].(string)
		if !ok {
			fmt.Println("Something went wrong for country: " + country)
		}

		// Create log in DB
		s := store.NewStore()
		url, err := s.GetUrlByID(urlId)
		if err != nil {
			return err
		}

		urlidUUID, err := gocql.ParseUUID(urlId)
		if err != nil {
			return err
		}
		urlLog := &types.UrlLog{UrlID: urlidUUID, ClientIP: clientIP, VisitedAt: time.Now(), City: city, Country: country}
		resp, err := http.Get(url.LongUrl)
		if err != nil {
			urlLog.HttpStatusCode = http.StatusInternalServerError
			urlLog.RedirectStatus = err.Error()
		} else {
			urlLog.HttpStatusCode = resp.StatusCode
			urlLog.RedirectStatus = resp.Status
		}

		s.CreateUrlLog(urlLog)
		return nil
	})

	if err := srv.Run(mux); err != nil {
		log.Fatalf("Asynq server error: %v", err)
	}
}
