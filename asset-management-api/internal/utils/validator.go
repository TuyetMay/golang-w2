package utils

import (
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
	
	// Register custom UUID validator
	validate.RegisterValidation("uuid", validateUUID)
}

func validateUUID(fl validator.FieldLevel) bool {
	_, err := uuid.Parse(fl.Field().String())
	return err == nil
}

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func ValidateStruct(s interface{}) []ValidationError {
	var validationErrors []ValidationError
	
	err := validate.Struct(s)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			var message string
			
			switch err.Tag() {
			case "required":
				message = "This field is required"
			case "email":
				message = "Invalid email format"
			case "min":
				message = "Value is too short"
			case "max":
				message = "Value is too long"
			case "uuid":
				message = "Invalid UUID format"
			case "oneof":
				message = "Invalid value. Allowed values: " + err.Param()
			default:
				message = "Invalid value"
			}
			
			validationErrors = append(validationErrors, ValidationError{
				Field:   strings.ToLower(err.Field()),
				Message: message,
			})
		}
	}
	
	return validationErrors
}

func GetValidationErrorMessages(errors []ValidationError) []string {
	var messages []string
	for _, err := range errors {
		messages = append(messages, err.Field+": "+err.Message)
	}
	return messages
}

// Common validation functions
func IsValidUUID(s string) bool {
	_, err := uuid.Parse(s)
	return err == nil
}

func IsValidAccessLevel(level string) bool {
	return level == "read" || level == "write"
}

func IsValidRole(role string) bool {
	return role == "manager" || role == "member"
}