package cmd

import (
	"testing"
)

func TestSplitter(t *testing.T) {
	t.Run("Command and argument with spaces should split in two", func(t *testing.T) {
		s := "ls -la"
		expected := []string{"ls", "-la"}
		result, _ := splitter(s)
		if len(result) != len(expected) {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})

	t.Run("Command and argument that has a string should split in two", func(t *testing.T) {
		s := "echo \"test\""
		expected := []string{"echo", "test"}
		result, _ := splitter(s)
		if len(result) != len(expected) {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})

	t.Run("Command and argument that has a string with spaces should split in two", func(t *testing.T) {
		s := "echo \"test test\""
		expected := []string{"echo", "test test"}
		result, _ := splitter(s)
		if len(result) != len(expected) {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})
}
