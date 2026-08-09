package testhelpers

import (
	"encoding/json"
	"fmt"
	"testing"
)

func AssertNoError(t *testing.T, err error, data any) {
	t.Helper()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	t.Logf("data: %s", formatData(data))
}

func formatData(data any) string {
	if data == nil {
		return "<nil>"
	}

	switch v := data.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	}

	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Sprintf("%v", data)
	}

	return string(jsonBytes)
}
