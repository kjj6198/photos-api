package utils

import (
	"github.com/satori/go.uuid"
)

// GenerateUUID creates uuid and ignore error.
func GenerateUUID() string {
	id, _ := uuid.NewV4()
	return id.String()
}
