package api

import (
	"context"
	"encoding/json"
	"errors"
	"io"
)

// StreamReader provides a streaming interface to the API
type StreamReader struct {
	events <-chan StreamEvent
	ctx    context.Context
}

// NewStreamReader creates a new stream reader
func NewStreamReader(ctx context.Context, events <-chan StreamEvent) *StreamReader {
	return &StreamReader{
		events: events,
		ctx:    ctx,
	}
}

// Next returns the next event from the stream
func (s *StreamReader) Next() (*StreamEvent, error) {
	select {
	case <-s.ctx.Done():
		return nil, s.ctx.Err()
	case event, ok := <-s.events:
		if !ok {
			return nil, io.EOF
		}
		return &event, nil
	}
}

// MessageStream provides a higher-level streaming interface
type MessageStream struct {
	reader *StreamReader
}

// NewMessageStream creates a new message stream
func NewMessageStream(ctx context.Context, events <-chan StreamEvent) *MessageStream {
	return &MessageStream{
		reader: NewStreamReader(ctx, events),
	}
}

// Message returns the complete message when ready
func (ms *MessageStream) Message() (*MessageResponse, error) {
	var msg MessageResponse
	var content []ContentBlock
	var usage MessageUsage
	var stopReason string

	for {
		event, err := ms.reader.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}

		switch event.Type {
		case "message_start":
			if data, ok := event.Data.(map[string]any); ok {
				if msgData, ok := data["message"].(map[string]any); ok {
					msg.Id = getString(msgData, "id")
					msg.Role = getString(msgData, "role")
					msg.Model = getString(msgData, "model")
					msg.Type = getString(msgData, "type")
				}
			}

		case "content_block_start":
			if data, ok := event.Data.(map[string]any); ok {
				if cbData, ok := data["content_block"].(map[string]any); ok {
					cb := ContentBlock{
						Type: getString(cbData, "type"),
						Text: getString(cbData, "text"),
					}
					content = append(content, cb)
				}
			}

		case "content_block_delta":
			if data, ok := event.Data.(map[string]any); ok {
				index := int(getFloat(data["index"]))
				if delta, ok := data["delta"].(map[string]any); ok {
					text := getString(delta, "text")
					if index < len(content) {
						content[index].Text += text
					}
				}
			}

		case "message_delta":
			if data, ok := event.Data.(map[string]any); ok {
				if delta, ok := data["delta"].(map[string]any); ok {
					stopReason = getString(delta, "stop_reason")
				}
				if usageData, ok := data["usage"].(map[string]any); ok {
					usage.InputTokens = int(getFloat(usageData["input_tokens"]))
					usage.OutputTokens = int(getFloat(usageData["output_tokens"]))
				}
			}

		case "message_stop":
			break

		case "error":
			return nil, errors.New(getString(event.Data.(map[string]any), "error"))
		}
	}

	msg.Content = content
	msg.StopReason = stopReason
	msg.Usage = usage
	return &msg, nil
}

// StreamChunks returns chunks as they arrive
func (ms *MessageStream) StreamChunks() (<-chan string, <-chan error) {
	textCh := make(chan string, 100)
	errCh := make(chan error, 1)

	go func() {
		defer close(textCh)
		defer close(errCh)

		for {
			event, err := ms.reader.Next()
			if err != nil {
				if !errors.Is(err, io.EOF) {
					errCh <- err
				}
				return
			}

			switch event.Type {
			case "content_block_delta":
				if data, ok := event.Data.(map[string]any); ok {
					if delta, ok := data["delta"].(map[string]any); ok {
						if text := getString(delta, "text"); text != "" {
							textCh <- text
						}
					}
				}

			case "message_stop":
				return

			case "error":
				if errMsg, ok := event.Data.(map[string]any); ok {
					errCh <- errors.New(getString(errMsg, "error"))
				}
				return
			}
		}
	}()

	return textCh, errCh
}

// Helper functions

func getString(m map[string]any, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func getFloat(v any) float64 {
	if f, ok := v.(float64); ok {
		return f
	}
	return 0
}

// ParseStreamEvent parses a raw JSON event
func ParseStreamEvent(data []byte) (*StreamEvent, error) {
	var event StreamEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return nil, err
	}
	return &event, nil
}
