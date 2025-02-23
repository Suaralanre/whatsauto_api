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
	logger *slog.Logger
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

type WhatsappButtonMessage struct {
	From string `json:"from"`
	ID string `json:"id"`
	Type string `json:"type"`
	Button struct {
		Payload string `json:"payload"`
		Text string `json:"text"`
	} `json:"button"`
	Context struct {
		From string `json:"from"`
		ID string `json:"id"`
	} `json:"context"`
}