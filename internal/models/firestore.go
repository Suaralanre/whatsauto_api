package models

import "context"

func (f *FirestoreClient) SaveAppointment(ctx context.Context, phoneNumber, eventID, subject string) error {
	client := f.Client

	_, err := client.Collection("appointments").Doc(phoneNumber).Set(ctx, Appointment{
			phoneNumber,
			eventID,
	})
	if err != nil {
		f.Logger.Error(err.Error(), "message", "unable to store Appointment details: phone, eventID")
		return err
	}
	
	f.Logger.Info("firestore save event","message", "Appointment successfully saved")
	return nil
}

func (f *FirestoreClient) GetAppointment(ctx context.Context, phoneNumber string) (Appointment, error) {
	client := f.Client

	doc, err := client.Collection("appointments").Doc(phoneNumber).Get(ctx)
	if err != nil {
		f.Logger.Error(err.Error(), "message", "unable to get Appointment details: phoneNumber")
		return Appointment{}, err
	}

	var appointment Appointment
	doc.DataTo(&appointment)

	f.Logger.Info("firestore get event","message", "Appointment successfully retrieved")
	return appointment, nil
}


func (f *FirestoreClient) DeleteAppointment(phoneNumber string) error {
	client := f.Client

	_, err := client.Collection("appointments").Doc(phoneNumber).Delete(context.Background())
	if err != nil {
		f.Logger.Error(err.Error(), "message", "unable to delete Appointment details: phone")
		return err
	}
	
	f.Logger.Info("firestore delete event","message", "Appointment successfully deleted")
	return nil
}