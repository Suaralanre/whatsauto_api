package main

import (
	"bytes"
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
	secret := utils.GetEnv("SECRET", "")

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

func (w *WhatsAppSender) sendAppointmentMessage(number, template, imageURL, procedure, day string) error {
	payload := map[string]interface{}{
		"messaging_product": "whatsapp",
		"to":                number,
		"recipient_type":    "individual",
		"type":              "template",
		"template": map[string]interface{}{
			"name": template,
			"language": map[string]string{
				"code":   "en",
				"policy": "deterministic",
			},
			"components": []interface{}{
				// Header (Image)
				map[string]interface{}{
					"type": "header",
					"parameters": []interface{}{
						map[string]interface{}{
							"type": "image",
							"image": map[string]string{
								"link": imageURL,
							},
						},
					},
				},
				// Body (Text Parameters)
				map[string]interface{}{
					"type": "body",
					"parameters": []interface{}{
						map[string]interface{}{
							"type":           "text",
							"parameter_name": "procedure",
							"text":           procedure,
						},
						map[string]interface{}{
							"type":           "text",
							"parameter_name": "day",
							"text":           day,
						},
					},
				},
			},
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("Error Serializing payload: %v", err)
	}

	req, err := http.NewRequest("POST", w.APIURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("Error creating request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+w.Token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Error sending request: %v", err.Error())
	}
	defer resp.Body.Close()

	status := resp.StatusCode

	if status == http.StatusOK {
		return nil
	}
	return fmt.Errorf("Error sending whatsapp message: %v", http.StatusInternalServerError)
}

func (w *WhatsAppSender) sendWelcomeMessage(number, template, imageURL, name, title string) error {
	payload := map[string]interface{}{
		"messaging_product": "whatsapp",
		"to":                number,
		"recipient_type":    "individual",
		"type":              "template",
		"template": map[string]interface{}{
			"name": template,
			"language": map[string]string{
				"code":   "en",
				"policy": "deterministic",
			},
			"components": []interface{}{
				// Header (Image)
				map[string]interface{}{
					"type": "header",
					"parameters": []interface{}{
						map[string]interface{}{
							"type": "image",
							"image": map[string]string{
								"link": imageURL,
							},
						},
					},
				},
				// Body (Text Parameters)
				map[string]interface{}{
					"type": "body",
					"parameters": []interface{}{
						map[string]interface{}{
							"type":           "text",
							"parameter_name": "title",
							"text":           title,
						},
						map[string]interface{}{
							"type":           "text",
							"parameter_name": "name",
							"text":           name,
						},
					},
				},
			},
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("Error Serializing payload: %v", err.Error())
	}

	req, err := http.NewRequest("POST", w.APIURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("Error creating request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+w.Token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Error sending request: %v", err.Error())
	}
	defer resp.Body.Close()

	status := resp.StatusCode

	if status == http.StatusOK {
		return nil
	}
	return fmt.Errorf("Error sending whatsapp message: %v", err)
}

func (w *WhatsAppSender) replyWhatsappMessage(number, body, messageID string) error {
	payload := map[string]interface{}{
		"messaging_product": "whatsapp",
		"to":                number,
		"type":              "text",
		"text": map[string]interface{}{
			"body":        body,
			"preview_url": false,
		},
		"context": map[string]interface{}{
			"message_id": messageID,
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("Error Serializing payload: %v", err.Error())
	}

	req, err := http.NewRequest("POST", w.APIURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("Error creating request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+w.Token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Error sending request: %v", err.Error())
	}
	defer resp.Body.Close()

	status := resp.StatusCode

	if status == http.StatusOK {
		w.logger.Info("Whatsapp message sent successfully")
		return nil
	}
	return fmt.Errorf("Error sending whatsapp message: %v", err)
}

