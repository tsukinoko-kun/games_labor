package ai

import (
	"fmt"
	"io"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/iterator"
)

// GenAIResponseReader implements io.Reader for a *genai.GenerateContentResponseIterator.
// It assumes the iterator yields responses containing only genai.Text parts.
// It expects exactly one candidate and one part per response.
type GenAIResponseReader struct {
	iter   *genai.GenerateContentResponseIterator
	buffer []byte // Holds leftover data from the previous read that didn't fit in p
}

// NewGenAIResponseReader creates a new GenAIResponseReader.
func NewGenAIResponseReader(
	iter *genai.GenerateContentResponseIterator,
) *GenAIResponseReader {
	return &GenAIResponseReader{
		iter: iter,
	}
}

// Read implements the io.Reader interface.
// It reads the text content from the underlying *genai.GenerateContentResponseIterator.
func (r *GenAIResponseReader) Read(p []byte) (n int, err error) {
	// If the internal buffer has data, satisfy the read from it first.
	if len(r.buffer) > 0 {
		n = copy(p, r.buffer)
		r.buffer = r.buffer[n:] // Advance the buffer
		return n, nil
	}

	// Buffer is empty, get the next response from the iterator.
	resp, err := r.iter.Next()
	if err != nil {
		if err == iterator.Done {
			return 0, io.EOF // Standard way to signal end of stream
		}
		// Wrap the underlying error for context.
		return 0, fmt.Errorf("error fetching next response from iterator: %w", err)
	}

	// Basic validation based on the expectation of Text content.
	if len(resp.Candidates) == 0 ||
		resp.Candidates[0].Content == nil ||
		len(resp.Candidates[0].Content.Parts) == 0 {
		// This case indicates an unexpected response structure (e.g., safety block).
		// Return 0 bytes read and nil error to allow the caller to retry.
		// Alternatively, could return an error if this state is considered fatal.
		return 0, nil
		// return 0, fmt.Errorf("received response with no processable content parts")
	}

	// Extract the text part, assuming it's the first and only part.
	part := resp.Candidates[0].Content.Parts[0]
	textPart, ok := part.(genai.Text)
	if !ok {
		// The response part was not of the expected genai.Text type.
		return 0, fmt.Errorf("expected genai.Text part, but got type %T", part)
	}

	// Convert the text to bytes.
	data := []byte(textPart)

	// Copy data into the provided buffer p.
	n = copy(p, data)

	// If the data read from the iterator was larger than the buffer p,
	// store the remainder in the internal buffer for the next Read call.
	if n < len(data) {
		r.buffer = data[n:]
	}

	// Return the number of bytes copied and nil error.
	return n, nil
}
