package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"
)

func TestPrintJSONMap(t *testing.T) {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	data := map[string]string{"key": "value"}
	err := printJSON(data)

	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("printJSON() error: %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)

	var parsed map[string]string
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if parsed["key"] != "value" {
		t.Errorf("parsed[key] = %q, want %q", parsed["key"], "value")
	}
}

func TestPrintJSONSlice(t *testing.T) {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := printJSON([]string{"a", "b", "c"})

	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("printJSON() error: %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)

	var parsed []string
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if len(parsed) != 3 {
		t.Errorf("expected 3 items, got %d", len(parsed))
	}
}

func TestPrintJSONNil(t *testing.T) {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := printJSON(nil)

	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("printJSON() error: %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	if buf.Len() == 0 {
		t.Error("expected some output for nil")
	}
}
