package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"
)

func CustomLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
	}))
}

// Parses the subject based on the way it was filled in Outlook:
// It is filled as: [0]: Whatsapp , [1]: Procedure, [2]: Patient Name
func ParseEventSubject(subject string) (string, string, string) {
	if !strings.HasPrefix(subject, "+") {
		return "", "", ""
	}
	parts := strings.Split(subject, ",")
	if len(parts) >= 3 {
		return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]), strings.TrimSpace((parts[2]))
	}
	return "", "", ""
}

// Parses the start and end time from outlook
func ParseDateTime(datetimeStr, timezone string) (string, error) {
	// Parse the time without timezone first
	parsedTime, err := time.Parse("2006-01-02T15:04:05.0000000", datetimeStr)
	if err != nil {
		return "", fmt.Errorf("Error parsing event time: %v", err)
	}

	// Load the location from the timezone string
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return "", fmt.Errorf("Error loading location: %v", err)
	}

	// Convert the time to the specified timezone
	parsedTime = parsedTime.In(loc)

	return parsedTime.Format("02 Jan 2006, 03:04 PM"), nil
}

// Utility function to send whatsapp message

// SendWhatsappMessage sends a WhatsApp message based on the type and additional arguments. arg[whatsapp message type, message, phonenumberID, ]
func SendWhatsappMessage(whatsappNumber, message string, arg ...string) error {
	// Default values
	businessName := ""
	// phone_number_id := arg[2]
	apiUrl := "https://graph.facebook.com/v17.0/YOUR_PHONE_NUMBER_ID/messages"
	apiToken := "YOUR_META_API_TOKEN"
	// var logger *slog.Logger

	// Determine the message type
	var msgType string
	if len(arg) > 0 {
		msgType = arg[0]
	} else {
		msgType = "text" // Default to text message
	}

	// Prepare the payload based on the message type
	var payload map[string]interface{}
	switch msgType {
	case "text":
		payload = map[string]interface{}{
			"messaging_product": "whatsapp",
			"to":                whatsappNumber,
			"type":              "text",
			"text": map[string]interface{}{
				"body": message,
			},
		}

	case "cta_url":
		if len(arg) < 3 {
			return fmt.Errorf("missing arguments for cta_url: display_text and url are required")
		}
		displayText := arg[1]
		url := arg[2]
		payload = map[string]interface{}{
			"messaging_product": "whatsapp",
			"recipient_type":    "individual",
			"to":                whatsappNumber,
			"type":              "interactive",
			"interactive": map[string]interface{}{
				"type": "cta_url",
				"header": map[string]interface{}{
					"text": businessName,
				},
				"body": map[string]interface{}{
					"text": message,
				},
				"footer": map[string]interface{}{
					"text": "",
				},
				"action": map[string]interface{}{
					"name": "cta_url",
					"parameters": map[string]interface{}{
						"display_text": displayText,
						"url":          url,
					},
				},
			},
		}

	case "list":
		if len(arg) < 2 {
			return fmt.Errorf("missing arguments for list: sections are required")
		}
		sections := arg[1] // JSON string representing sections
		var sectionsData []map[string]interface{}
		if err := json.Unmarshal([]byte(sections), &sectionsData); err != nil {
			return fmt.Errorf("invalid sections JSON: %v", err)
		}
		payload = map[string]interface{}{
			"messaging_product": "whatsapp",
			"to":                whatsappNumber,
			"type":              "interactive",
			"interactive": map[string]interface{}{
				"type": "list",
				"header": map[string]interface{}{
					"type": "text",
					"text": businessName,
				},
				"body": map[string]interface{}{
					"text": message,
				},
				"action": map[string]interface{}{
					"button":   "Choose an Option",
					"sections": sectionsData,
				},
			},
		}

	default:
		return fmt.Errorf("unsupported message type: %s", msgType)
	}

	// Convert payload to JSON
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %v", err)
	}

	// Send the request to the WhatsApp API
	req, err := http.NewRequest("POST", apiUrl, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+apiToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("WhatsApp API error: %s, response: %s", resp.Status, string(body))
	}

	return nil
}

func SendTemplateWhatsappMsg(whatsappNumber string, message string, arg ...string) {

}
