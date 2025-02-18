package models

import (
	"context"
	"fmt"
	"log"
	"log/slog"

	"cloud.google.com/go/firestore"
	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/Suaralanre/whatsauto_api/internal/utils"
)

type FirestoreClient struct {
	client *firestore.Client
	logger *slog.Logger
}

type Appointment struct {
	Phonenumber string   `firestore:"phone_number"`
	EventID     string   `firestore:"event_id"`
	Categories  []string `firestore:"categories"`
}

func (f *FirestoreClient) GetSecret(secretName string) ([]byte, error) {
	ctx := context.Background()
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create secret manager client: %v", err)
	}
	defer client.Close()
	projectID := utils.GetEnv("GOOGLE_PROJECT_ID", "")

	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: fmt.Sprintf("projects/%s/secrets/%s/versions/latest", projectID, secretName),
	}

	result, err := client.AccessSecretVersion(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to access secret: %v", err)
	}

	return result.Payload.Data, nil
}

func (f *FirestoreClient) LoadServiceAccount() ([]byte, error) {
	secretName := utils.GetEnv("SECRET_MANAGER", "")
	secretData, err := f.GetSecret(secretName)
	if err != nil {
		return nil, err
	}
	return secretData, nil
}

func (f *FirestoreClient) GetFirestoreClient() (*firestore.Client, error) {
	projectID := utils.GetEnv("GOOGLE_PROJECT_ID", "")
	ctx := context.Background()
	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Error creating Firestore client: %v", err)
	}
	return client, nil
}
