package format

import (
	"fmt"
	"reflect"
	"time"
)

// OutputValidator provides validation for command outputs
type OutputValidator struct {
	schemas map[string]Schema
}

// Schema defines the validation rules for a specific command type
type Schema struct {
	RequiredFields []string
	FieldTypes     map[string]reflect.Kind
	CustomRules    []ValidationRule
}

// ValidationRule defines a custom validation function
type ValidationRule func(data map[string]interface{}) error

// NewOutputValidator creates a new validator with default schemas
func NewOutputValidator() *OutputValidator {
	validator := &OutputValidator{
		schemas: make(map[string]Schema),
	}

	// Register default schemas
	validator.registerDefaultSchemas()

	return validator
}

// registerDefaultSchemas registers validation schemas for common command types
func (v *OutputValidator) registerDefaultSchemas() {
	// Base schema that all outputs must follow
	baseSchema := Schema{
		RequiredFields: []string{"type", "status", "timestamp", "data"},
		FieldTypes: map[string]reflect.Kind{
			"type":      reflect.String,
			"status":    reflect.String,
			"timestamp": reflect.String,
			"data":      reflect.Map,
		},
		CustomRules: []ValidationRule{
			v.validateStatus,
			v.validateTimestamp,
		},
	}

	// Auth command schema
	authSchema := baseSchema
	authSchema.CustomRules = append(authSchema.CustomRules, v.validateAuthData)
	v.schemas["auth"] = authSchema

	// Error output schema (less strict than auth)
	errorSchema := baseSchema
	v.schemas["error"] = errorSchema

	// Chain command schema
	chainSchema := baseSchema
	chainSchema.CustomRules = append(chainSchema.CustomRules, v.validateChainData)
	v.schemas["chain"] = chainSchema

	// Transaction command schema
	transactionSchema := baseSchema
	transactionSchema.CustomRules = append(transactionSchema.CustomRules, v.validateTransactionData)
	v.schemas["transaction"] = transactionSchema

	// Wallet address command schema
	walletAddressSchema := baseSchema
	walletAddressSchema.CustomRules = append(walletAddressSchema.CustomRules, v.validateWalletAddressData)
	v.schemas["wallet_address"] = walletAddressSchema

	// Wallet balance command schema
	walletBalanceSchema := baseSchema
	walletBalanceSchema.CustomRules = append(walletBalanceSchema.CustomRules, v.validateWalletBalanceData)
	v.schemas["wallet_balance"] = walletBalanceSchema

	// Default schema for unknown command types
	v.schemas["default"] = baseSchema
}

// ValidateJSON validates a JSON output against the appropriate schema
func (v *OutputValidator) ValidateJSON(data map[string]interface{}) error {
	// Get command type
	commandType, ok := data["type"].(string)
	if !ok {
		return fmt.Errorf("missing or invalid 'type' field")
	}

	// Get schema for command type
	schema, exists := v.schemas[commandType]
	if !exists {
		schema = v.schemas["default"]
	}

	return v.ValidateSchema(commandType, data, schema)
}

// ValidateSchema validates data against a specific schema
func (v *OutputValidator) ValidateSchema(commandType string, data map[string]interface{}, schema Schema) error {
	// Check required fields
	for _, field := range schema.RequiredFields {
		if _, exists := data[field]; !exists {
			return fmt.Errorf("missing required field '%s' for command type '%s'", field, commandType)
		}
	}

	// Check field types
	for field, expectedType := range schema.FieldTypes {
		if value, exists := data[field]; exists {
			actualType := reflect.TypeOf(value).Kind()
			if actualType != expectedType {
				return fmt.Errorf("field '%s' has type %v, expected %v", field, actualType, expectedType)
			}
		}
	}

	// Apply custom validation rules
	for _, rule := range schema.CustomRules {
		if err := rule(data); err != nil {
			return fmt.Errorf("validation rule failed for command type '%s': %w", commandType, err)
		}
	}

	return nil
}

// validateStatus ensures status field has valid values
func (v *OutputValidator) validateStatus(data map[string]interface{}) error {
	status, ok := data["status"].(string)
	if !ok {
		return fmt.Errorf("status field must be a string")
	}

	if status != "success" && status != "error" {
		return fmt.Errorf("status must be 'success' or 'error', got: %s", status)
	}

	return nil
}

// validateTimestamp ensures timestamp field is in valid ISO 8601 format
func (v *OutputValidator) validateTimestamp(data map[string]interface{}) error {
	timestamp, ok := data["timestamp"].(string)
	if !ok {
		return fmt.Errorf("timestamp field must be a string")
	}

	if timestamp == "" {
		return fmt.Errorf("timestamp cannot be empty")
	}

	// Validate ISO 8601 format
	if _, err := time.Parse(time.RFC3339, timestamp); err != nil {
		return fmt.Errorf("timestamp must be in ISO 8601 format (RFC3339): %w", err)
	}

	return nil
}

// validateAuthData validates authentication command data
func (v *OutputValidator) validateAuthData(data map[string]interface{}) error {
	dataField, ok := data["data"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("data field must be an object for auth commands")
	}

	// Check for required auth fields
	requiredFields := []string{"node", "username"}
	for _, field := range requiredFields {
		if _, exists := dataField[field]; !exists {
			return fmt.Errorf("auth data missing required field: %s", field)
		}
	}

	// Validate field types
	if node, exists := dataField["node"]; exists {
		if _, ok := node.(string); !ok {
			return fmt.Errorf("auth data 'node' field must be a string")
		}
	}

	if username, exists := dataField["username"]; exists {
		if _, ok := username.(string); !ok {
			return fmt.Errorf("auth data 'username' field must be a string")
		}
	}

	return nil
}

// validateChainData validates chain command data
func (v *OutputValidator) validateChainData(data map[string]interface{}) error {
	dataField, ok := data["data"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("data field must be an object for chain commands")
	}

	// Chain-specific validation can be added here
	// For now, just ensure it's a valid object
	_ = dataField

	return nil
}

// validateTransactionData validates transaction command data
func (v *OutputValidator) validateTransactionData(data map[string]interface{}) error {
	dataField, ok := data["data"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("data field must be an object for transaction commands")
	}

	// Transaction-specific validation can be added here
	// For now, just ensure it's a valid object
	_ = dataField

	return nil
}

// validateWalletAddressData validates wallet address command data
func (v *OutputValidator) validateWalletAddressData(data map[string]interface{}) error {
	dataField, ok := data["data"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("data field must be an object for wallet address commands")
	}

	// Check for required fields
	if _, exists := dataField["addressIndex"]; !exists {
		return fmt.Errorf("wallet address data must contain 'addressIndex' field")
	}

	// For success status, address must be present and non-empty
	status, _ := data["status"].(string)
	if status == "success" {
		address, exists := dataField["address"].(string)
		if !exists || address == "" {
			return fmt.Errorf("wallet address data must contain non-empty 'address' field for success status")
		}
	}

	return nil
}

// validateWalletBalanceData validates wallet balance command data
func (v *OutputValidator) validateWalletBalanceData(data map[string]interface{}) error {
	dataField, ok := data["data"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("data field must be an object for wallet balance commands")
	}

	// Check for required fields
	if _, exists := dataField["addressIndex"]; !exists {
		return fmt.Errorf("wallet balance data must contain 'addressIndex' field")
	}

	// For success status, address and balances must be present
	status, _ := data["status"].(string)
	if status == "success" {
		address, exists := dataField["address"].(string)
		if !exists || address == "" {
			return fmt.Errorf("wallet balance data must contain non-empty 'address' field for success status")
		}

		if _, exists := dataField["balances"]; !exists {
			return fmt.Errorf("wallet balance data must contain 'balances' field for success status")
		}
	}

	return nil
}

// RegisterSchema allows registering custom schemas for new command types
func (v *OutputValidator) RegisterSchema(commandType string, schema Schema) {
	v.schemas[commandType] = schema
}

// AddValidationRule adds a custom validation rule to an existing schema
func (v *OutputValidator) AddValidationRule(commandType string, rule ValidationRule) error {
	schema, exists := v.schemas[commandType]
	if !exists {
		return fmt.Errorf("schema for command type '%s' does not exist", commandType)
	}

	schema.CustomRules = append(schema.CustomRules, rule)
	v.schemas[commandType] = schema

	return nil
}
