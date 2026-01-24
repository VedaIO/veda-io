package write

// WriteRequest represents a request to write to the database.
type WriteRequest struct {
	Query string
	Args  []interface{}
}

// writeCh is the internal channel for sequencing write requests.
var writeCh = make(chan WriteRequest, 100)

// EnqueueWrite sends a write request to the database writer channel.
func EnqueueWrite(query string, args ...interface{}) {
	writeCh <- WriteRequest{Query: query, Args: args}
}

// GetWriteChannel returns the internal write channel for the writer to consume.
func GetWriteChannel() chan WriteRequest {
	return writeCh
}
