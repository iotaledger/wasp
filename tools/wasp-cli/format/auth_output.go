package format

import (
	"fmt"
)

// AuthData represents the data structure for authentication commands
type AuthData struct {
	Node     string `json:"node"`
	Username string `json:"username"`
	Message  string `json:"message,omitempty"`
}

// AuthOutput represents the output of authentication commands
type AuthOutput struct {
	BaseOutput[AuthData]
}

// NewAuthOutput creates a new authentication output
func NewAuthOutput(node, username, message string, success bool) *AuthOutput {
	status := "success"
	if !success {
		status = "error"
	}

	data := AuthData{
		Node:     node,
		Username: username,
		Message:  message,
	}

	return &AuthOutput{
		BaseOutput: NewBaseOutput("auth", status, data),
	}
}

// NewAuthSuccess creates a successful authentication output
func NewAuthSuccess(node, username string) *AuthOutput {
	return NewAuthOutput(node, username, "Authentication successful", true)
}

// NewAuthError creates an error authentication output
func NewAuthError(node, username, errorMsg string) *AuthOutput {
	return NewAuthOutput(node, username, errorMsg, false)
}

// ToTable returns the authentication output as table rows
func (ao *AuthOutput) ToTable() [][]string {
	rows := [][]string{
		{"Status", "Node", "Username"},
		{ao.Status, ao.Data.Node, ao.Data.Username},
	}

	// Add message row if present
	if ao.Data.Message != "" {
		rows[0] = append(rows[0], "Message")
		rows[1] = append(rows[1], ao.Data.Message)
	}

	return rows
}

// Validate validates the authentication output
func (ao *AuthOutput) Validate() error {
	// Validate base output
	if err := ao.BaseOutput.Validate(); err != nil {
		return err
	}

	// Validate auth-specific fields
	if ao.Data.Node == "" {
		return fmt.Errorf("node cannot be empty for auth output")
	}

	if ao.Data.Username == "" {
		return fmt.Errorf("username cannot be empty for auth output")
	}

	// For error status, message should not be empty
	if ao.Status == "error" && ao.Data.Message == "" {
		return fmt.Errorf("error message cannot be empty for failed auth output")
	}

	return nil
}

// GetNode returns the node identifier
func (ao *AuthOutput) GetNode() string {
	return ao.Data.Node
}

// GetUsername returns the username
func (ao *AuthOutput) GetUsername() string {
	return ao.Data.Username
}

// GetMessage returns the message
func (ao *AuthOutput) GetMessage() string {
	return ao.Data.Message
}

// IsSuccess returns true if the authentication was successful
func (ao *AuthOutput) IsSuccess() bool {
	return ao.Status == "success"
}
