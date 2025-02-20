package models

import "context"

func (f *FirestoreClient) SaveAppointment(phoneNumber, eventID string) error {
	client := f.Client

	defer client.Close()

	_, err := client.Collection("appointments").Doc(phoneNumber).Set(context.Background(), Appointment{
			phoneNumber,
			eventID,
	})
	if err != nil {
		f.Logger.Error(err.Error(), "message", "unable to store Appointment details: phone, eventID")
	}
	
	f.Logger.Info("firestore save event","message", "Appointment successfully saved")
	return nil
}