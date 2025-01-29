package utils

import (
	"fmt"
	"log/slog"
	"os"
)

func CustomLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
	}))
}

// Utility function to send whatsapp message.
func SendWhatsappMessage(whatsappNumber, message string, arg ...string) {
	// apiurl, apitoken, type, display_text, url
	businessName := ""
	apiUrl := ""
	apiToken := ""

	payload :=  map[string]interface{}{
		"messaging_product": "whatsapp",
		"to": whatsappNumber,
		"type": "interactive",
		"interactive": map[string]interface{}{
			"type": arg[0],
			"header": map[string]interface{}{
				"text": fmt.Sprintf("Welcome to %s!", businessName),
			},
			"body": map[string]interface{}{
				"text": "Click the button below to visit our social media pages.",
			},
			"footer": map[string]interface{}{
				"text": "We look forward to serving you",
			},
			"action": map[string]interface{}{
				"buttons": []map[string]interface{}{
					"type": arg[0]
				}
			},
		},
	}
}