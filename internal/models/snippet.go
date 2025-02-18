package models

import "context"

func (f *FirestoreClient) SaveAppointment(phoneNumber, eventID string, categories []string) error {
	client, err := f.GetFirestoreClient()
	if err != nil {
		f.logger.Error(err.Error(), "message", "Unable to get firestore client")
		return err
	}
	defer client.Close()

	_, err = client.Collection("appointments").Doc(phoneNumber).Set(context.Background(), Appointment{
			phoneNumber,
			eventID,
			categories,
	})
	if err != nil {
		f.logger.Error(err.Error(), "message", "Error storing appointment")
	}
	f.logger.Info("firestore save event","message", "Appointment successfully saved")
	return nil
}