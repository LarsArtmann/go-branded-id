//go:build !goexperiment.jsonv2

package id

import (
	"encoding/json/v2"
	"fmt"
)

func marshalJSON(v any) ([]byte, error) {
	return json.Marshal(v)
}

func unmarshalJSON(data []byte, v any) error {
	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("unmarshal json: %w", err)
	}

	return nil
}
