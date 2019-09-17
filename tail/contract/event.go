package contract

import "fmt"

// GSEvent is the payload of a GCS event.
type GSEvent struct {
	Bucket string `json:"bucket"`
	Name   string `json:"name"`
}

func (e *GSEvent) URL() string {
	return fmt.Sprintf("gs://%v/%v", e.Bucket, e.Name)
}
