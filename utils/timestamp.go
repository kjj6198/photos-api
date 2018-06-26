package utils

import (
	"encoding/json"
	"time"
)

type TimeStamp time.Time

func (t TimeStamp) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Now().Unix())
}
