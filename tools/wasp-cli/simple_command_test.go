package main

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

// TestSimpleCommandExecution demonstrates basic cobra command testing
// This test doesn't require any external dependencies and shows the core concept
func TestSimpleCommandExecution(t *testing.T) {
	// Create a simple test command
	var output bytes.Buffer

	rootCmd := &cobra.Command{
		Use:   "wasp-cli",
		Short: "Test wasp-cli",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Println("wasp-cli test output")
			return nil
		},
	}

	// Add a simple subcommand
	walletCmd := &cobra.Command{
		Use:   "wallet",
		Short: "Wallet operations",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Println("wallet command executed")
			return nil
		},
	}

	rootCmd.AddCommand(walletCmd)
	rootCmd.SetOut(&output)
	rootCmd.SetErr(&output)

	t.Run("root command", func(t *testing.T) {
		output.Reset()
		rootCmd.SetArgs([]string{})

		err := rootCmd.Execute()
		require.NoError(t, err)

		result := output.String()
		require.Contains(t, result, "wasp-cli test output")
	})

	t.Run("wallet subcommand", func(t *testing.T) {
		output.Reset()
		rootCmd.SetArgs([]string{"wallet"})

		err := rootCmd.Execute()
		require.NoError(t, err)

		result := output.String()
		require.Contains(t, result, "wallet command executed")
	})

	t.Run("help command", func(t *testing.T) {
		output.Reset()
		rootCmd.SetArgs([]string{"--help"})

		err := rootCmd.Execute()
		require.NoError(t, err)

		result := output.String()
		require.Contains(t, result, "wasp-cli")
		require.Contains(t, result, "Usage:")
	})

	t.Run("invalid command", func(t *testing.T) {
		output.Reset()
		rootCmd.SetArgs([]string{"invalid"})

		err := rootCmd.Execute()
		require.Error(t, err)
		require.Contains(t, err.Error(), "unknown command")
	})
}

// TestBasicMockConcept demonstrates the basic mocking concept
func TestBasicMockConcept(t *testing.T) {
	// Simple mock structure
	type MockService struct {
		Value  string
		Called bool
	}

	mock := &MockService{Value: "test-value"}

	// Test command that uses the mock
	var output bytes.Buffer
	cmd := &cobra.Command{
		Use: "test",
		RunE: func(cmd *cobra.Command, args []string) error {
			mock.Called = true
			cmd.Printf("Service returned: %s\n", mock.Value)
			return nil
		},
	}

	cmd.SetOut(&output)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.NoError(t, err)
	require.True(t, mock.Called)
	require.Contains(t, output.String(), "test-value")
}
