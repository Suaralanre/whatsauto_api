package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)
	router.HandlerFunc(http.MethodPost, "/v1/patients/new", app.NewPatientFormHandler)
	router.HandlerFunc(http.MethodGet, "/v1/patients/appointments", app.CalendarEventsHandler)
	router.HandlerFunc(http.MethodGet, "/v1/whatsapp/webhook", app.WhatsappWebhookInitializer)
	router.HandlerFunc(http.MethodPost, "/v1/whatsapp/webhook", app.WhatsappWebhookHandler)

	return router
}
