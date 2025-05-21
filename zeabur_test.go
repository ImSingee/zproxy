package main

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

var (
	testApiKey   = os.Getenv("ZEABUR_API_KEY")
	testServerId = os.Getenv("ZEABUR_SERVER_ID")
)

func TestBuildZeaburDnsMap(t *testing.T) {
	// Skip the test if environment variables are not set
	if testApiKey == "" || testServerId == "" {
		t.Skip("ZEABUR_API_KEY or ZEABUR_SERVER_ID environment variables not set")
	}

	dnsMap, err := buildZeaburDnsMap(testApiKey, testServerId)
	if err != nil {
		t.Fatal(err)
	}
	for k, v := range dnsMap {
		fmt.Printf("%s\t%s\n", k, v)
	}
}

func TestBuildZeaburDnsMapErrors(t *testing.T) {
	// Test cases for error scenarios
	testCases := []struct {
		name        string
		apiKey      string
		serverId    string
		expectedErr string
		setup       func() // Function to set up the test environment
		teardown    func() // Function to clean up after the test
	}{
		{
			name:        "Empty API Key",
			apiKey:      "",
			serverId:    "server-123",
			expectedErr: "API request failed with status",
			setup:       func() {},
			teardown:    func() {},
		},
		{
			name:        "Empty Server ID",
			apiKey:      testApiKey,
			serverId:    "",
			expectedErr: "API request failed with status",
			setup:       func() {},
			teardown:    func() {},
		},
		{
			name:        "Invalid Server ID",
			apiKey:      testApiKey,
			serverId:    "server-nonexistent",
			expectedErr: "API request failed with status",
			setup:       func() {},
			teardown:    func() {},
		},
	}

	// Run each test case
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set up the test environment
			tc.setup()

			// Call the function with test case parameters
			_, err := buildZeaburDnsMap(tc.apiKey, tc.serverId)

			// Clean up after the test
			tc.teardown()

			// Check if the error is as expected
			if err == nil {
				t.Fatalf("Expected error containing '%s', but got nil", tc.expectedErr)
			}
			if !strings.Contains(err.Error(), tc.expectedErr) {
				t.Fatalf("Expected error containing '%s', but got '%s'", tc.expectedErr, err.Error())
			}
		})
	}
}
