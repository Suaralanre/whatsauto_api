package main

import (
	"context"
	"os"

	"cloud.google.com/go/firestore"
	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
)

func (app *application) GetServiceAccountKey() ([]byte, error) {
	ctx := context.Background()
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		app.logger.Error(err.Error(), "message", "failed to create secret manager client")
		return nil, err
	}
	defer client.Close()

	secretName := "projects/whatsapp-automation-451109/secrets/service-account-key/versions/latest"

	req := &secretmanagerpb.AccessSecretVersionRequest{Name: secretName}
	result, err := client.AccessSecretVersion(ctx, req)
	if err != nil {
		app.logger.Error(err.Error(), "message", "failed to access secret version")
		return nil, err
	}
	return result.Payload.Data, nil

}

func (app *application) LoadGoogleCredentials() error {
	keyData, err := app.GetServiceAccountKey()
	if err != nil {
		return err
	}

	tempFile, err := os.CreateTemp("", "gcp-key-*.json")
	if err != nil {
		return err
	}
	defer tempFile.Close()

	_, err = tempFile.Write(keyData)
	if err != nil {
		return err
	}

	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", tempFile.Name())
	return nil
}

func NewFirestoreClient(projectID string) (*FirestoreClient, error) {
	ctx := context.Background()
	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		return nil, err
	}

	return &FirestoreClient{client: client}, nil
}
