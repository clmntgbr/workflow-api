package validation

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
)

var ErrValidationFailed = errors.New("validation failed")

var validate *validator.Validate

func init() {
	validate = validator.New()
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" || name == "" {
			return fld.Name
		}
		return name
	})
	_ = validate.RegisterValidation("workflow_status", validateWorkflowStatus)
	_ = validate.RegisterValidation("endpoint_status", validateEndpointStatus)
	_ = validate.RegisterValidation("http_method", validateHTTPMethod)
	_ = validate.RegisterValidation("schedule_type", validateScheduleType)
	_ = validate.RegisterValidation("schedule_unit", validateScheduleUnit)
}

func validateWorkflowStatus(fl validator.FieldLevel) bool {
	switch fl.Field().String() {
	case "active", "inactive", "canceled":
		return true
	default:
		return false
	}
}

func validateEndpointStatus(fl validator.FieldLevel) bool {
	switch fl.Field().String() {
	case "active", "inactive":
		return true
	default:
		return false
	}
}

func validateHTTPMethod(fl validator.FieldLevel) bool {
	switch strings.ToUpper(fl.Field().String()) {
	case "GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS":
		return true
	default:
		return false
	}
}

func validateScheduleType(fl validator.FieldLevel) bool {
	switch fl.Field().String() {
	case "none", "recurring", "once":
		return true
	default:
		return false
	}
}

func validateScheduleUnit(fl validator.FieldLevel) bool {
	switch fl.Field().String() {
	case "minute", "hour", "day", "week", "month", "year":
		return true
	default:
		return false
	}
}

func BindBody(c fiber.Ctx, dst any) error {
	if err := c.Bind().Body(dst); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
		})
	}
	return Struct(c, dst)
}

func Struct(c fiber.Ctx, dst any) error {
	if err := validate.Struct(dst); err != nil {
		_ = c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Validation failed",
			"errors":  formatErrors(err),
		})
		return ErrValidationFailed
	}
	return nil
}

// FiberErrorHandler suppresses ErrValidationFailed because the response is already written.
func FiberErrorHandler(c fiber.Ctx, err error) error {
	if errors.Is(err, ErrValidationFailed) {
		return nil
	}
	return fiber.DefaultErrorHandler(c, err)
}

func formatErrors(err error) map[string]string {
	out := make(map[string]string)
	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		out["_error"] = err.Error()
		return out
	}

	for _, fieldErr := range validationErrors {
		out[fieldErr.Field()] = messageFor(fieldErr)
	}
	return out
}

func messageFor(fieldErr validator.FieldError) string {
	switch fieldErr.Tag() {
	case "required":
		return "is required"
	case "uuid":
		return "must be a valid UUID"
	case "url":
		return "must be a valid URL"
	case "min":
		if fieldErr.Kind() == reflect.String {
			return fmt.Sprintf("must be at least %s characters", fieldErr.Param())
		}
		return fmt.Sprintf("must be at least %s", fieldErr.Param())
	case "max":
		if fieldErr.Kind() == reflect.String {
			return fmt.Sprintf("must be at most %s characters", fieldErr.Param())
		}
		return fmt.Sprintf("must be at most %s", fieldErr.Param())
	case "oneof", "workflow_status", "endpoint_status", "http_method", "schedule_type", "schedule_unit":
		return "is invalid"
	default:
		return fmt.Sprintf("failed on '%s' validation", fieldErr.Tag())
	}
}
