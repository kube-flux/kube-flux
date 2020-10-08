package policy

import (
	"testing"
)

func TestNewPolicyHandler(t *testing.T) {
	handler, err := NewPolicyHandler()
	if handler == nil {
		t.Fatalf("Error %v", err)
	}
}
