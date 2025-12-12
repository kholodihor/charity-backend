package api

import "fmt"

const minDonationAmount = 100

type createGoalRequest struct {
	Title        string  `json:"title"`
	Description  *string `json:"description"`
	TargetAmount int64   `json:"target_amount"`
}

type updateGoalRequest struct {
	Title        *string `json:"title"`
	Description  *string `json:"description"`
	TargetAmount *int64  `json:"target_amount"`
	IsActive     *bool   `json:"is_active"`
}

type createDonationRequest struct {
	UserID      int64  `json:"user_id"`
	GoalID      int64  `json:"goal_id"`
	Amount      int64  `json:"amount"`
	Currency    string `json:"currency"`
	IsAnonymous bool   `json:"is_anonymous"`
}

type createUserRequest struct {
	Email    string  `json:"email"`
	Name     *string `json:"name"`
	Password string  `json:"password"`
}

func validateCreateUserRequest(req createUserRequest) error {
	if req.Email == "" {
		return fmt.Errorf("email is required")
	}
	if req.Password == "" {
		return fmt.Errorf("password is required")
	}
	return nil
}

type loginUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func validateLoginUserRequest(req loginUserRequest) error {
	if req.Email == "" {
		return fmt.Errorf("email is required")
	}
	if req.Password == "" {
		return fmt.Errorf("password is required")
	}
	return nil
}

func validateCreateGoalRequest(req createGoalRequest) error {
	if req.Title == "" {
		return fmt.Errorf("title is required")
	}
	if req.TargetAmount <= 0 {
		return fmt.Errorf("target_amount must be positive")
	}
	return nil
}

func validateUpdateGoalRequest(req updateGoalRequest) error {
	// Require at least one field to update
	if req.Title == nil && req.Description == nil && req.TargetAmount == nil && req.IsActive == nil {
		return fmt.Errorf("no fields to update")
	}

	if req.TargetAmount != nil && *req.TargetAmount <= 0 {
		return fmt.Errorf("target_amount must be positive")
	}
	return nil
}

func validateCreateDonationRequest(req createDonationRequest) error {
	if req.GoalID <= 0 {
		return fmt.Errorf("goal_id must be positive")
	}

	if req.Amount < minDonationAmount {
		return fmt.Errorf("amount must be at least %d", minDonationAmount)
	}
	if req.Currency == "" {
		return fmt.Errorf("currency is required")
	}
	return nil
}
