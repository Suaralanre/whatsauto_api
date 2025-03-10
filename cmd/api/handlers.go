package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Suaralanre/whatsauto_api/internal/utils"
	"github.com/Suaralanre/whatsauto_api/internal/validator"
)

var confirmed string = "Thank you for confirming your appointment. We look forward to seeing you soon."
var cancelled string = "Your appointment has been cancelled. Kindly place a call or send a whatsapp message to the phone number in the original message to let us know when you will be available."

// Handler for new patient's form submission
func (app *application) NewPatientFormHandler(w http.ResponseWriter, r *http.Request) {
	var input PatientForm

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.errorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	v := validator.New()

	v.Check(input.Title != "", "title", "must be provided")
	v.Check(input.FirstName != "", "firstname", "must be provided")
	v.Check(input.Whatsapp != "", "whatsapp", "must be provided")
	v.Check(strings.HasPrefix(input.Whatsapp, "+"), "whatsapp", "Whatsapp number must start with a country code.")

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	image := utils.GetEnv("IMAGE_URL", "")
	template := utils.GetEnv("WELCOME_TEMPLATE", "")

	// Send whatsapp message to phone number
	err = app.sender.sendWelcomeMessage(input.Whatsapp, template, image, input.FirstName, input.Title)
	if err != nil {
		app.logger.Error(err.Error(), "message", "error sending welcome message")
	}
}

// Handler for Calendar appointments
func (app *application) CalendarEventsHandler(w http.ResponseWriter, r *http.Request) {

	now := time.Now()
	dayAfterTomorrow := now.AddDate(0, 0, 1)
	dayStr := dayAfterTomorrow.Format("2006-01-02")

	startDateTime := url.QueryEscape(fmt.Sprintf("%sT00:00:00Z", dayStr))
	endDateTime := url.QueryEscape(fmt.Sprintf("%sT23:59:59Z", dayStr))

	userEmail := utils.GetEnv("OUTLOOK_EMAIL", "")
	imageURL := utils.GetEnv("IMAGE_URL", "")
	template := utils.GetEnv("APPOINTMENT_TEMPLATE", "")

	url := fmt.Sprintf(
		"https://graph.microsoft.com/v1.0/users/%s/calendarView?startDateTime=%s&endDateTime=%s&$top=30&$select=start,end,subject,categories",
		userEmail, startDateTime, endDateTime,
	)
	var events []Event

	for {
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			app.logError(r, err)
			return
		}

		req.Header.Add("Authorization", "Bearer "+*app.outlook)
		req.Header.Add("Accept", "application/json")
		req.Header.Add("Prefer", "outlook.timezone=\"Africa/Lagos\"")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			app.logError(r, err)
			return
		}

		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			app.logger.Error(err.Error(), "message", "Error parsing Body")
			return
		}

		if resp.StatusCode != http.StatusOK {
			app.logger.Info("Outlook Issue", "Message", "Unable to get outlook events")
			return
		}

		var result Result

		// could not use readJSON here because resp is of type http.Response
		if err := json.Unmarshal(body, &result); err != nil {
			app.logger.Error(err.Error(), "message", "Error parsing body")
		}
		// Append the events from this page to all "events"
		events = append(events, result.Value...)

		if len(events) >= 30 || result.NextPageLink == "" {
			break
		}

		// Update the URL to fetch the next page
		url = result.NextPageLink
	}

	for _, event := range events {
		subject := strings.TrimSpace(event.Subject)
		if strings.HasPrefix(subject, "+") {
			whatsapp, procedure, _ := utils.ParseEventSubject(subject)
			startTime, _ := utils.ParseDateTime(event.Start.DateTime, event.Start.Timezone)

			// add event_id, phonenumber to firestore
			err := app.firestore.SaveAppointment(r.Context(), whatsapp, event.ID, subject)
			if err != nil {
				app.logger.Error(err.Error())
			}

			// send whatsapp message
			err = app.sender.sendAppointmentMessage(whatsapp, template, imageURL, procedure, startTime)
			if err != nil {
				app.logger.Error(err.Error(), "message", "error sending appointment message")
			}
			if err == nil {
				app.logger.Info("WhatsApp message sent successfully", "number", whatsapp)
			}
		} else {
			app.logger.Info(subject, "message", "Event subject not well formatted")
			continue
		}
	}
	return
}

func (app *application) WhatsappWebhookInitializer(w http.ResponseWriter, r *http.Request) {
	token := utils.GetEnv("WHATSAPP_WEBHOOK_TOKEN", "secret")
	mode := r.URL.Query().Get("hub.mode")
	challenge := r.URL.Query().Get("hub.challenge")
	verifyToken := r.URL.Query().Get("hub.verify_token")

	if mode == "subscribe" && verifyToken == token {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(challenge)); err != nil {
			app.logger.Error(err.Error(), "message", "Unable to verify token hook")
			return
		}
		app.logger.Info("200", "message", "Whatsapp hook verified")
	}
	return
}

func (app *application) WhatsappWebhookHandler(w http.ResponseWriter, r *http.Request) {

	userID := utils.GetEnv("OUTLOOK_EMAIL", "")
	// Read and log raw payload
	body, err := io.ReadAll(r.Body)
	if err != nil {
		app.logger.Error(err.Error(), "message", "Error reading webhook body")
	}
	// fmt.Printf("Raw webhook payload: %s\n", string(body))

	// Decode the payload
	var payload WhatsappButtonMessage
	if err := json.Unmarshal(body, &payload); err != nil {
		app.logger.Error(err.Error(), "message", "Error decoding webhook response")
		return
	}
	// fmt.Printf("Decoded webhook payload: %+v\n", payload)

	// Extract the button click details
	if len(payload.Entry) > 0 &&
		len(payload.Entry[0].Changes) > 0 &&
		len(payload.Entry[0].Changes[0].Value.Messages) > 0 &&
		payload.Entry[0].Changes[0].Value.Messages[0].Type == "button" {

		message := payload.Entry[0].Changes[0].Value.Messages[0]
		whatsapp := "+" + message.From
		// fmt.Printf("Phone number: %s, Button clicked: %s\n", whatsapp, message.Button.Text)

		// Retrieve appointment from Firestore
		event, err := app.firestore.GetAppointment(r.Context(), whatsapp)
		if err != nil {
			app.logger.Error(err.Error(), "message", "Error getting appointment")
			return
		}

		// Check if Confirm or Cancel button was clicked
		if message.Button.Text == "Cancel" {
			// Update Outlook event category to "NO SHOW"
			err = app.changeOutlookCategory(event.EventID, userID)
			if err != nil {
				app.logger.Error(err.Error(), "message", "Error changing outlook category")
				return
			}

			// Send Cancel confirmation on WhatsApp
			if err = app.sender.replyWhatsappMessage(whatsapp, cancelled, message.ID); err != nil {
				app.logger.Error(err.Error(), "message", "Error sending cancel whatsapp message")
				return
			}

		} else if message.Button.Text == "Confirm" {
			// Add a tick to outlook event subject
			if err = app.changeOutlookSubject(event.EventID, userID); err != nil {
				app.logger.Error(err.Error(), "message", "Error changing outlook subject")
			}
			// Send confirmation reply
			if err = app.sender.replyWhatsappMessage(whatsapp, confirmed, message.ID); err != nil {
				app.logger.Error(err.Error(), "message", "Error sending confirm whatsapp message")
				return
			}
		}

		// Delete appointment from Firestore after handling the response
		if err = app.firestore.DeleteAppointment(whatsapp); err != nil {
			app.logger.Error(err.Error(), "message", "Error deleting appointment")
			return
		}
	} else {
		app.logger.Info("Webhook received but payload missing expected message data.")
	}
}
