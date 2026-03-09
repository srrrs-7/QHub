package consulting

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// SSEWriter writes Server-Sent Events to an http.ResponseWriter.
// It requires the underlying writer to implement http.Flusher.
type SSEWriter struct {
	w       http.ResponseWriter
	flusher http.Flusher
}

// NewSSEWriter creates an SSEWriter from the given ResponseWriter.
// Returns an error if the writer does not support flushing.
func NewSSEWriter(w http.ResponseWriter) (*SSEWriter, error) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return nil, fmt.Errorf("streaming not supported: ResponseWriter does not implement http.Flusher")
	}
	return &SSEWriter{w: w, flusher: flusher}, nil
}

// WriteEvent writes a named SSE event with the given data string.
func (s *SSEWriter) WriteEvent(event, data string) error {
	_, err := fmt.Fprintf(s.w, "event: %s\ndata: %s\n\n", event, data)
	if err != nil {
		return err
	}
	s.flusher.Flush()
	return nil
}

// WriteMessage writes an SSE "message" event with the given messageResponse serialized as JSON.
func (s *SSEWriter) WriteMessage(msg messageResponse) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return s.WriteEvent("message", string(data))
}

// WriteDone writes an SSE "done" event indicating the stream has completed.
func (s *SSEWriter) WriteDone() error {
	return s.WriteEvent("done", `{"status":"complete"}`)
}

// WriteError writes an SSE "error" event with the given error message.
func (s *SSEWriter) WriteError(err error) error {
	data, _ := json.Marshal(map[string]string{"error": err.Error()})
	return s.WriteEvent("error", string(data))
}

