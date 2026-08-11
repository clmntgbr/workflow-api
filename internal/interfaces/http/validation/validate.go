package validation

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
)

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
}

func validateWorkflowStatus(fl validator.FieldLevel) bool {
	switch fl.Field().String() {
	case "active", "inactive", "canceled":
		return true
	default:
		return false
	}
}

// BindBody binds the request body then validates the destination struct.
func BindBody(c fiber.Ctx, dst any) error {
	if err := c.Bind().Body(dst); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
		})
	}
	return Struct(c, dst)
}

// Struct validates an already-bound payload.
func Struct(c fiber.Ctx, dst any) error {
	if err := validate.Struct(dst); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Validation failed",
			"errors":  formatErrors(err),
		})
	}
	return nil
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
	case "oneof", "workflow_status":
		return "is invalid"
	default:
		return fmt.Sprintf("failed on '%s' validation", fieldErr.Tag())
	}
}
