package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/Suaralanre/whatsauto_api/internal/models"
	"github.com/Suaralanre/whatsauto_api/internal/utils"
)


func main() {
	// environmental variables
	logger := utils.CustomLogger()
	port := utils.GetEnvInt("PORT", 8080)
	env := utils.GetEnv("ENVIRONMENT", "development")
	phoneID := utils.GetEnv("PHONE_NUMBER_ID", "")
	apiUrl := fmt.Sprintf("https://graph.facebook.com/v21.0/%s/messages", phoneID)
	tokenWA := utils.GetEnv("WHATSAPP_TOKEN", "")

	// firestore initialization
	firestoreClient := &models.FirestoreClient{Logger: logger}
	client, err := firestoreClient.GetFirestoreClient()
	if err != nil{
		firestoreClient.Logger.Error(err.Error(), "message", "Unable to initialize firestore client")
	}
	firestoreClient.Client = client
	
	// application struct initialization
	app := &application{
		logger: logger,
		sender: &WhatsAppSender{
				apiUrl,
				tokenWA,
				logger,
		},
		firestore: 
		firestoreClient,
	}
	
	// outlook initialization
	token, err := app.getOutlookAccessToken()
	if err != nil{
		app.logger.Error(err.Error(), "message", "outlook authentication error")
	}
	app.outlook = &token


	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  10 * time.Second,
		ErrorLog:     slog.NewLogLogger(logger.Handler(), slog.LevelError),
	}

	logger.Info("starting server", "addr", srv.Addr, "env", env)
	err = srv.ListenAndServe()
	app.logger.Error(err.Error())
	os.Exit(1)
}
