package policy

import (
	"testing"
)

func TestNewPolicyHandler(t *testing.T) {
	handler, err := NewPolicyHandler()
	if handler == nil {
		t.Fatalf("Error %v", err)
	}

	// TODO: Use hooks to test
}

// TODO(lacee): Test if GET request will function as expected

// TODO: Test if POST & DELETE requests will fail
