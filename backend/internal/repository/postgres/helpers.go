package postgres

import (
	"encoding/json"
)

func encodeJSON(v any) []byte {
	if v == nil {
		return []byte("{}")
	}
	b, err := json.Marshal(v)
	if err != nil {
		return []byte("{}")
	}
	return b
}

func decodeJSONMap(b []byte) map[string]any {
	if len(b) == 0 {
		return map[string]any{}
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		return map[string]any{}
	}
	return m
}

func decodeJSONSlice(b []byte) []any {
	if len(b) == 0 {
		return nil
	}
	var s []any
	_ = json.Unmarshal(b, &s)
	return s
}
