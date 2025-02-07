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

	// Send whatsapp message to phone number
	fmt.Fprintf(w, "New Patient: %s %s sent to number %s", input.Title, input.FirstName, input.Whatsapp)
}

func (app *application) CalendarEvents(w http.ResponseWriter, r *http.Request) {

	accessToken, err := app.getOutlookAccessToken()
	if err != nil {
		app.logger.Error(err.Error(), "message", "Error getting Access")
		return
	}

	now := time.Now()
	dayAfterTomorrow := now.AddDate(0, 0, 2)
	dayStr := dayAfterTomorrow.Format("2006-01-02")

	startDateTime := url.QueryEscape(fmt.Sprintf("%sT00:00:00Z", dayStr))
	endDateTime := url.QueryEscape(fmt.Sprintf("%sT23:59:59Z", dayStr))

	userEmail := utils.GetEnv("OUTLOOK_EMAIL", "")

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

		req.Header.Add("Authorization", "Bearer "+accessToken)
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

	for _, event := range events{
		if strings.HasPrefix(event.Subject, "+") {
			// send whatsapp message
			// get response
			// parse response 
			// change category
		}
	}
}
