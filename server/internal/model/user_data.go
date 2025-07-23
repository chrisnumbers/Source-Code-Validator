package model

type UserData struct {
	Url              string `bson:"url"`
	RequirementsData string `bson:"requirements_data"`
	Consultation     string `bson:"consultation"`
}
