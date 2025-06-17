package inbox

import (
	"encoding/json"

	"github.com/google/uuid"
)

type Message struct {
	ID        uuid.UUID
	EventType string
	Payload   json.RawMessage
}
