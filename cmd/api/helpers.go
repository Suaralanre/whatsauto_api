package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/confidential"
	"github.com/Suaralanre/whatsauto_api/internal/utils"
)

type envelope map[string]any

func (app *application) writeJSON(w http.ResponseWriter, status int, data any, headers http.Header) error {
	js, err := json.Marshal(data)
	if err != nil {
		return err
	}

	js = append(js, '\n')

	for key, value := range headers {
		w.Header()[key] = value
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)

	return nil
}

func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	maxBytes := 2_048_576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	err := dec.Decode(dst)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError
		var maxBytesError *http.MaxBytesError

		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly formed JSON (at character %d)", syntaxError.Offset)

		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly formed JSON")

		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains an invalid value for the %q field", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)

		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")

		case strings.HasPrefix(err.Error(), "json: unknown field"):
			fieldName := strings.TrimPrefix(err.Error(), "json:unknown field ")
			return fmt.Errorf("body contains unknown key %s", fieldName)

		case errors.As(err, &maxBytesError):
			return fmt.Errorf("body must not be larger than %d bytes", maxBytesError.Limit)

		case errors.As(err, &invalidUnmarshalError):
			panic(err)

		default:
			return err
		}
	}
	err = dec.Decode(&struct{}{})
	if !errors.Is(err, io.EOF) {
		return errors.New("body must only contain a single JSON value")
	}
	return nil
}

func (app *application) getOutlookAccessToken() (string, error) {
	tenant_id := utils.GetEnv("TENANT_ID", "")
	client_id := utils.GetEnv("CLIENT_ID", "")
	secret := utils.GetEnv("CLIENT_ID", "")

	cred, err := confidential.NewCredFromSecret(secret)
	if err != nil {
		app.logger.Error(err.Error(), "message", "Error with secret")
		return "", err
	}
	confidentialClient, err := confidential.New(fmt.Sprintf("https://login.microsoftonline.com/%s", tenant_id), client_id, cred)

	scopes := []string{"https://graph.microsoft.com/.default"}
	result, err := confidentialClient.AcquireTokenSilent(context.TODO(), scopes)
	if err != nil {
		// cache miss, authenticate with another AcquireToken... method
		result, err = confidentialClient.AcquireTokenByCredential(context.TODO(), scopes)
		if err != nil {
			app.logger.Error(err.Error(), "message", "Error with tenantID or clientID")
			return "", err
		}
	}
	return result.AccessToken, nil
}

// func (app *application) fetchCalendarEvents(accessToken string) ([]Event, error) {
// 	now := time.Now()
// 	dayAfterTomorrow := now.AddDate(0, 0, 2)
// 	dayStr := dayAfterTomorrow.Format("2006-01-02")

// 	startDateTime := url.QueryEscape(fmt.Sprintf("%sT00:00:00Z", dayStr))
// 	endDateTime := url.QueryEscape(fmt.Sprintf("%sT23:59:59Z", dayStr))

// 	userEmail := utils.GetEnv("OUTLOOK_EMAIL", "")

// 	url := fmt.Sprintf(
// 		"https://graph.microsoft.com/v1.0/users/%s/calendarView?startDateTime=%s&endDateTime=%s&$top=30&$select=start,end,subject,categories",
// 		userEmail, startDateTime, endDateTime,
// 	)

// 	var allEvents []Event

// 	for {
// 		req, err := http.NewRequest(http.MethodGet, url, nil)
// 		if err != nil {
// 			app.logError(req, err)
// 			return nil, err
// 		}

// 		req.Header.Add("Authorization", "Bearer "+accessToken)
// 		req.Header.Add("Accept", "application/json")
// 		req.Header.Add("Prefer", "outlook.timezone=\"Africa/Lagos\"")

// 		client := &http.Client{}
// 		resp, err := client.Do(req)
// 		if err != nil {
// 			return nil, err
// 		}

// 		defer resp.Body.Close()

// 		body, err := io.ReadAll(resp.Body)
// 		if err != nil {

// 		}
// 	}
// }
