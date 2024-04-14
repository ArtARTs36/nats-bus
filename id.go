package natsbus

import "github.com/google/uuid"

const messageHeaderMessageID = "message-id"

func generateMessageID() string {
	return uuid.New().String()
}
