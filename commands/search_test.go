package commands

import (
	"testing"
	"fmt"
)

func TestGetNextLoadingChar(t *testing.T) {
	char := getNextLoadingChar()
	if char != "/" {
		t.Errorf("Expected /, but got %s", char)
	}

	char = getNextLoadingChar()
	if char != "-" {
		t.Errorf("Expected -, but got %s", char)
	}

	char = getNextLoadingChar()
	if char != "\\" {
		t.Errorf("Expected \\, but got %s", char)
	}

	char = getNextLoadingChar()
	if char != "|" {
		t.Errorf("Expected |, but got %s", char)
	}

	char = getNextLoadingChar()
	if char != "/" {
		t.Errorf("Expected /, but got %s", char)
	}
}
