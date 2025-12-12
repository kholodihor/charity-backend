package token

import (
	"time"
)

// Maker defines the behavior for creating and verifying tokens.
type Maker interface {
	// CreateToken creates a new token for a specific user, role and duration.
	CreateToken(name string, role string, duration time.Duration, tokenType TokenType) (string, *Payload, error)
	// VerifyToken checks if the token is valid or not.
	VerifyToken(token string, tokenType TokenType) (*Payload, error)
}
