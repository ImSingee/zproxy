package main

import (
	"bytes"
	"io"
	"os"
	"testing"
)

func TestZeaburDnsStore_Set(t *testing.T) {
	// Create a new store
	store := NewZeaburDnsStore()

	// Capture stdout to verify printed output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Test case 1: Adding new entries to an empty store
	newMap1 := map[string]string{
		"service1.example.com": "10.0.0.1",
		"service2.example.com": "10.0.0.2",
	}
	store.Set(newMap1)

	// Close writer and restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read captured output
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output1 := buf.String()

	// Verify output contains additions
	expectedAdditions := []string{
		"+ service1.example.com \t -> \t 10.0.0.1",
		"+ service2.example.com \t -> \t 10.0.0.2",
	}
	for _, expected := range expectedAdditions {
		if !contains(output1, expected) {
			t.Errorf("Expected output to contain '%s', but it didn't. Output: %s", expected, output1)
		}
	}

	// Test case 2: Modifying and removing entries
	// Capture stdout again
	r, w, _ = os.Pipe()
	os.Stdout = w

	newMap2 := map[string]string{
		"service1.example.com": "10.0.0.3", // Changed value
		"service3.example.com": "10.0.0.4", // New entry
		// service2.example.com is removed
	}
	store.Set(newMap2)

	// Close writer and restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read captured output
	buf.Reset()
	io.Copy(&buf, r)
	output2 := buf.String()

	// Verify output contains modifications and removals
	expectedChanges := []string{
		"- service1.example.com \t -> \t 10.0.0.1", // Old value removed
		"+ service1.example.com \t -> \t 10.0.0.3", // New value added
		"- service2.example.com \t -> \t 10.0.0.2", // Entry removed
		"+ service3.example.com \t -> \t 10.0.0.4", // New entry added
	}
	for _, expected := range expectedChanges {
		if !contains(output2, expected) {
			t.Errorf("Expected output to contain '%s', but it didn't. Output: %s", expected, output2)
		}
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return bytes.Contains([]byte(s), []byte(substr))
}