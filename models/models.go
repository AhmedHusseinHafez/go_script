// Package models defines the data structures used across the workflow.
package models

// WorkflowState holds all intermediate state accumulated during the orchestration run.
type WorkflowState struct {
	// SP fields
	SPNationalID string
	SPEmail      string
	SPToken      string

	// Company
	CompanyID string

	// FP fields
	FPToken string

	// Funding
	FundingRequestID string

	// Opportunity
	OpportunityID string
}

// SPRegisterRequest represents the payload for SP registration.
type SPRegisterRequest struct {
	UserType             int    `json:"user_type"`
	Name                 string `json:"name"`
	Email                string `json:"email"`
	NationalID           string `json:"national_id"`
	PhoneNumber          string `json:"phone_number"`
	Address              string `json:"address"`
	Password             string `json:"password"`
	PasswordConfirmation string `json:"password_confirmation"`
}

// LoginRequest represents the payload for login.
type LoginRequest struct {
	NationalID string `json:"national_id,omitempty"`
	Email      string `json:"email,omitempty"`
	Password   string `json:"password,omitempty"`
	UserType   int    `json:"user_type,omitempty"`
}

// OTPRequest represents the payload for OTP verification.
type OTPRequest struct {
	Email string `json:"email"`
	OTP   string `json:"otp"`
}
