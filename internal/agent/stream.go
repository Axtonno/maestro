package agent

import (
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"

	pkgAgent "github.com/antonio-cafeo/maestro/pkg/agent"
	"github.com/antonio-cafeo/maestro/pkg/provider"
)

type partialToolCall struct {
	id        string
	name      string
	arguments strings.Builder
}

func assembleStream(stream provider.Stream, maxBytes, maxCalls int) (provider.CompletionResponse, error) {
	if stream == nil || typedNil(stream) || maxBytes <= 0 || maxCalls <= 0 {
		return provider.CompletionResponse{}, errors.New("invalid stream assembler input")
	}
	var content strings.Builder
	calls := make(map[int]*partialToolCall)
	response := provider.CompletionResponse{Message: provider.Message{Role: provider.RoleAssistant}}
	total := 0
	for {
		chunk, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			if closeErr := stream.Close(); closeErr != nil {
				return provider.CompletionResponse{}, closeErr
			}
			break
		}
		if err != nil {
			return provider.CompletionResponse{}, errors.Join(err, stream.Close())
		}
		content.WriteString(chunk.Content)
		total += len(chunk.Content)
		if chunk.Model != "" {
			response.Model = chunk.Model
		}
		if chunk.FinishReason != "" {
			response.FinishReason = chunk.FinishReason
		}
		if chunk.Usage.InputTokens > response.Usage.InputTokens {
			response.Usage.InputTokens = chunk.Usage.InputTokens
		}
		if chunk.Usage.OutputTokens > response.Usage.OutputTokens {
			response.Usage.OutputTokens = chunk.Usage.OutputTokens
		}
		for _, delta := range chunk.ToolCalls {
			if delta.Index < 0 {
				_ = stream.Close()
				return provider.CompletionResponse{}, errors.New("negative tool delta index")
			}
			partial, exists := calls[delta.Index]
			if !exists {
				if len(calls) >= maxCalls {
					_ = stream.Close()
					return provider.CompletionResponse{}, fmt.Errorf("stream tool call limit exceeded: %w", pkgAgent.ErrLimitExceeded)
				}
				partial = &partialToolCall{}
				calls[delta.Index] = partial
			}
			if delta.ID != "" {
				if partial.id != "" && partial.id != delta.ID {
					_ = stream.Close()
					return provider.CompletionResponse{}, errors.New("conflicting stream tool call ID")
				}
				partial.id = delta.ID
			}
			if delta.Name != "" {
				if partial.name != "" && partial.name != delta.Name {
					_ = stream.Close()
					return provider.CompletionResponse{}, errors.New("conflicting stream tool name")
				}
				partial.name = delta.Name
			}
			partial.arguments.WriteString(delta.Arguments)
			total += len(delta.ID) + len(delta.Name) + len(delta.Arguments)
		}
		if total > maxBytes {
			_ = stream.Close()
			return provider.CompletionResponse{}, fmt.Errorf("stream byte limit exceeded: %w", pkgAgent.ErrLimitExceeded)
		}
	}
	response.Message.Content = content.String()
	indices := make([]int, 0, len(calls))
	for index := range calls {
		indices = append(indices, index)
	}
	sort.Ints(indices)
	for _, index := range indices {
		partial := calls[index]
		response.Message.ToolCalls = append(response.Message.ToolCalls, provider.ToolCall{
			ID: partial.id, Name: partial.name, Arguments: []byte(partial.arguments.String()),
		})
	}
	return response, nil
}
