package data

import (
	"encoding/json"
	"fmt"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// convertProtoMessageToJSONMap converts a protobuf message to a map[string]any.
func convertProtoMessageToJSONMap(pb proto.Message) (map[string]any, error) {
	// Marshal the protobuf message to JSON
	jsonData, err := protojson.Marshal(pb)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal proto message to JSON: %w", err)
	}

	// Unmarshal the JSON data into a map
	var result map[string]any
	if err := json.Unmarshal(jsonData, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON to map: %w", err)
	}

	return result, nil
}
