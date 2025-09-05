package format

import (
	"fmt"
	"os"
	"runtime/debug"
	"strings"

	"github.com/iotaledger/wasp/v2/clients/apiextensions"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/log"
	"github.com/spf13/cobra"
)

// CommandErrorData represents the data structure for command errors
type CommandErrorData struct {
	Command     string
	Error       string
	ErrorType   string
	DetailError string `json:",omitempty"`
	Context     string `json:",omitempty"`
}

// CommandErrorOutput represents an error result from a cobra command
type CommandErrorOutput struct {
	BaseOutput[CommandErrorData]
}

// NewCommandErrorOutput creates a new command error output
func NewCommandErrorOutput(commandName, errorMsg, errorType string) *CommandErrorOutput {
	data := CommandErrorData{
		Command:   commandName,
		Error:     errorMsg,
		ErrorType: errorType,
	}

	return &CommandErrorOutput{
		BaseOutput: NewBaseOutput("command", "error", data),
	}
}

// NewCommandErrorFromError creates a command error output from a Go error
func NewCommandErrorFromError(commandName string, err error) *CommandErrorOutput {
	if err == nil {
		return nil
	}

	errorType := "standard"
	errorMsg := err.Error()
	detailError := ""

	// Check if it's an API error and extract additional information
	if apiError, ok := apiextensions.AsAPIError(err); ok {
		errorType = "api"
		if strings.Contains(apiError.Error, "401") {
			errorMsg = "unauthorized request: are you logged in? (wasp-cli login)"
		} else {
			errorMsg = apiError.Error
			if apiError.DetailError != nil {
				detailError = apiError.DetailError.Error + ": " + apiError.DetailError.Message
			}
		}
	}

	output := NewCommandErrorOutput(commandName, errorMsg, errorType)
	if detailError != "" {
		output.Data.DetailError = detailError
	}

	return output
}

// ToTable returns the command error output as table rows
func (ceo *CommandErrorOutput) ToTable() [][]string {
	rows := [][]string{
		{"Command", "Status", "Error Type", "Error"},
		{ceo.Data.Command, ceo.Status, ceo.Data.ErrorType, ceo.Data.Error},
	}

	// Add detail error row if present
	if ceo.Data.DetailError != "" {
		rows[0] = append(rows[0], "Details")
		rows[1] = append(rows[1], ceo.Data.DetailError)
	}

	// Add context row if present
	if ceo.Data.Context != "" {
		rows[0] = append(rows[0], "Context")
		rows[1] = append(rows[1], ceo.Data.Context)
	}

	return rows
}

// Validate validates the command error output
func (ceo *CommandErrorOutput) Validate() error {
	// Validate base output
	if err := ceo.BaseOutput.Validate(); err != nil {
		return err
	}

	// Validate command error specific fields
	if ceo.Data.Command == "" {
		return fmt.Errorf("command name cannot be empty for command error output")
	}

	if ceo.Data.Error == "" {
		return fmt.Errorf("error message cannot be empty for command error output")
	}

	if ceo.Data.ErrorType == "" {
		return fmt.Errorf("error type cannot be empty for command error output")
	}

	return nil
}

// WithContext adds context information to the error
func (ceo *CommandErrorOutput) WithContext(context string) *CommandErrorOutput {
	ceo.Data.Context = context
	return ceo
}

// GetCommand returns the command name
func (ceo *CommandErrorOutput) GetCommand() string {
	return ceo.Data.Command
}

// GetError returns the error message
func (ceo *CommandErrorOutput) GetError() string {
	return ceo.Data.Error
}

// GetErrorType returns the error type
func (ceo *CommandErrorOutput) GetErrorType() string {
	return ceo.Data.ErrorType
}

// FormatCommandError formats and prints a command error using the unified output system
func FormatCommandError(cmd *cobra.Command, err error) error {
	if err == nil {
		return nil
	}

	commandName := cmd.Name()
	if cmd.Parent() != nil {
		commandName = cmd.Parent().Name() + " " + commandName
	}

	errorOutput := NewCommandErrorFromError(commandName, err)

	// Use the existing formatter to print the error
	formatter := NewFormatter()
	return formatter.PrintOutput(errorOutput)
}

// WrapRunE wraps a function to be used with cobra.Command.RunE and handles errors using glazed
func WrapRunE(runFunc func(cmd *cobra.Command, args []string) error) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		err := runFunc(cmd, args)
		if err != nil {
			// Format and print the error using glazed
			if formatErr := FormatCommandError(cmd, err); formatErr != nil {
				// If formatting fails, fall back to the original log.Check behavior
				log.Check(err)
			}
			// Return the original error to maintain cobra's error handling behavior
			return err
		}
		return nil
	}
}

// FormatAndExitWithError formats an error using glazed and exits the program
// This function is called from main() to handle all command errors consistently
func FormatAndExitWithError(cmd *cobra.Command, err error) error {
	if err == nil {
		return nil
	}

	// Get the command name for error context
	commandName := getCommandName(cmd)

	// Create error output
	errorOutput := NewCommandErrorFromError(commandName, err)

	// Format and print the error
	formatter := NewFormatter()
	if formatErr := formatter.PrintOutput(errorOutput); formatErr != nil {
		// If formatting fails, return the format error so main() can fall back to log.Check
		return formatErr
	}

	// Exit the program after formatting the error
	if log.DebugFlag {
		fmt.Printf("%s", debug.Stack())
	}
	os.Exit(1)
	return nil // This line will never be reached, but satisfies the return type
}

// getCommandName extracts the full command name including parent commands
func getCommandName(cmd *cobra.Command) string {
	if cmd == nil {
		return "unknown"
	}

	var parts []string
	current := cmd
	for current != nil && current.Name() != "wasp-cli" {
		parts = append([]string{current.Name()}, parts...)
		current = current.Parent()
	}

	if len(parts) == 0 {
		return "wasp-cli"
	}

	return strings.Join(parts, " ")
}

// HandleCommandError is a convenience function for handling errors in RunE functions
// This function should NOT format errors - just return them to be handled by main()
func HandleCommandError(cmd *cobra.Command, err error) error {
	// Simply return the error - formatting will be done in main()
	return err
}
