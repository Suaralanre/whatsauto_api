package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/Suaralanre/whatsauto_api/internal/utils"
)


func main() {
	logger := utils.CustomLogger()
	port := utils.GetEnvInt("PORT", 8080)
	env := utils.GetEnv("ENVIRONMENT", "development")

	app := &application{
		logger: logger,
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  10 * time.Second,
		ErrorLog:     slog.NewLogLogger(logger.Handler(), slog.LevelError),
	}

	logger.Info("starting server", "addr", srv.Addr, "env", env)
	err := srv.ListenAndServe()
	logger.Error(err.Error())
	os.Exit(1)
}
