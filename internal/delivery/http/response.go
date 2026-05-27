package http

import (
	"encoding/json"
	"errors"
	nethttp "net/http"
	"strings"
	"unicode"

	"go-arch/internal/domain"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type apiResponse struct {
	Success   bool          `json:"success"`
	Message   string        `json:"message"`
	Data      any           `json:"data,omitempty"`
	Errors    []errorDetail `json:"errors,omitempty"`
	RequestID string        `json:"request_id,omitempty"`
}

type errorDetail struct {
	Field   string `json:"field,omitempty"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func respondSuccess(ctx *gin.Context, status int, message string, data any) {
	ctx.JSON(status, apiResponse{
		Success:   true,
		Message:   message,
		Data:      data,
		RequestID: requestIDFromContext(ctx),
	})
}

func respondFailure(ctx *gin.Context, status int, message string, details []errorDetail) {
	ctx.JSON(status, apiResponse{
		Success:   false,
		Message:   message,
		Errors:    details,
		RequestID: requestIDFromContext(ctx),
	})
}

func respondFailureWithData(ctx *gin.Context, status int, message string, data any, details []errorDetail) {
	ctx.JSON(status, apiResponse{
		Success:   false,
		Message:   message,
		Data:      data,
		Errors:    details,
		RequestID: requestIDFromContext(ctx),
	})
}

func respondValidationError(ctx *gin.Context, err error) {
	respondFailure(ctx, nethttp.StatusBadRequest, "request validation failed", validationErrorDetails(err))
}

func respondDomainError(ctx *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidInput):
		respondFailure(ctx, nethttp.StatusBadRequest, "invalid request", []errorDetail{
			{Code: "invalid_input", Message: err.Error()},
		})
	case errors.Is(err, domain.ErrUserNotFound):
		respondFailure(ctx, nethttp.StatusNotFound, "resource not found", []errorDetail{
			{Code: "user_not_found", Message: err.Error()},
		})
	case errors.Is(err, domain.ErrEmailTaken), errors.Is(err, domain.ErrDuplicateEmail):
		respondFailure(ctx, nethttp.StatusConflict, "resource conflict", []errorDetail{
			{Field: "email", Code: "email_taken", Message: "email is already registered"},
		})
	default:
		respondFailure(ctx, nethttp.StatusInternalServerError, "internal server error", []errorDetail{
			{Code: "internal_error", Message: "unexpected server error"},
		})
	}
}

func validationErrorDetails(err error) []errorDetail {
	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		details := make([]errorDetail, 0, len(validationErrors))
		for _, fieldError := range validationErrors {
			details = append(details, errorDetail{
				Field:   jsonFieldName(fieldError.Field()),
				Code:    fieldError.Tag(),
				Message: validationMessage(fieldError),
			})
		}

		return details
	}

	var syntaxError *json.SyntaxError
	if errors.As(err, &syntaxError) {
		return []errorDetail{
			{Code: "invalid_json", Message: "request body contains invalid JSON"},
		}
	}

	return []errorDetail{
		{Code: "invalid_request_body", Message: "request body is invalid"},
	}
}

func validationMessage(fieldError validator.FieldError) string {
	field := jsonFieldName(fieldError.Field())

	switch fieldError.Tag() {
	case "required":
		return field + " is required"
	case "email":
		return field + " must be a valid email address"
	case "min":
		return field + " must be at least " + fieldError.Param() + " characters"
	case "max":
		return field + " must be at most " + fieldError.Param() + " characters"
	default:
		return field + " is invalid"
	}
}

func jsonFieldName(field string) string {
	if field == "" {
		return ""
	}

	var builder strings.Builder
	for index, char := range field {
		if unicode.IsUpper(char) {
			if index > 0 {
				builder.WriteRune('_')
			}
			builder.WriteRune(unicode.ToLower(char))
			continue
		}

		builder.WriteRune(char)
	}

	return builder.String()
}
