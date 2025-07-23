package util

import "go.mongodb.org/mongo-driver/mongo"

type Handler struct {
	DB *mongo.Client
}

func NewHandler(client *mongo.Client) *Handler {
	return &Handler{DB: client}
}
