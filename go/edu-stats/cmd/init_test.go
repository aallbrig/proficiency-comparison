package cmd

import (
	"bytes"
	"testing"
)

func TestInitCommand(t *testing.T) {
	// Capture output
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	
	// Test that init command exists
	initCmd := rootCmd.Commands()
	found := false
	for _, cmd := range initCmd {
		if cmd.Name() == "init" {
			found = true
			
			// Verify command properties
			if cmd.Short == "" {
				t.Error("Init command should have a short description")
			}
			
			if cmd.Long == "" {
				t.Error("Init command should have a long description")
			}
			
			break
		}
	}
	
	if !found {
		t.Error("Init command not found in root commands")
	}
}

func TestInitCommandHelp(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"init", "--help"})
	
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Init help command failed: %v", err)
	}
	
	output := buf.String()
	
	// Check for key phrases in help text
	expectedPhrases := []string{
		"Initialize",
		"database",
		"schema.sql",
		"CREATE TABLE IF NOT EXISTS",
	}
	
	for _, phrase := range expectedPhrases {
		if !bytes.Contains([]byte(output), []byte(phrase)) {
			t.Errorf("Help text should contain '%s'", phrase)
		}
	}
}
