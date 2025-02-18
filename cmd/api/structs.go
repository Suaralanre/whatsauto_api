package main

import (
	"log/slog"

	"github.com/Suaralanre/whatsauto_api/internal/models"
)

type application struct {
	logger    *slog.Logger
	sender    *WhatsAppSender
	firestore *models.FirestoreClient
	outlook   *string
}

type WhatsAppSender struct {
	APIURL string
	Token  string
}

type PatientForm struct {
	Title     string `json:"title"`
	FirstName string `json:"firstname"`
	Whatsapp  string `json:"whatsapp"`
}

type EventTime struct {
	DateTime string `json:"dateTime"`
	Timezone string `json:"timeZone"`
}

type Event struct {
	ID         string    `json:"id"`
	Subject    string    `json:"subject"`
	Start      EventTime `json:"start"`
	End        EventTime `json:"end"`
	Categories []string  `json:"categories"`
}

type Result struct {
	Value        []Event `json:"value"`
	NextPageLink string  `json:"@odata.nextLink"`
}
