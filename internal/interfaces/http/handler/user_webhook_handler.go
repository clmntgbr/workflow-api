package handler

import (
	"encoding/json"
	"errors"
	"log"

	usercmd "go-api/internal/application/command/user"
	"go-api/internal/interfaces/http/dto"
	"go-api/internal/interfaces/http/validation"

	"github.com/gofiber/fiber/v3"
)

var errUserNotFound = errors.New("user not found")

type UserWebhookHandler struct {
	getUserByExternalIDHandler    userWebhookGetByExternalIDHandler
	createUserHandler             userWebhookCreateUserHandler
	updateUserHandler             userWebhookUpdateUserHandler
	deleteUserByExternalIDHandler userWebhookDeleteByExternalIDHandler
}

func NewUserWebhookHandler(
	getUserByExternalIDHandler userWebhookGetByExternalIDHandler,
	createUserHandler userWebhookCreateUserHandler,
	updateUserHandler userWebhookUpdateUserHandler,
	deleteUserByExternalIDHandler userWebhookDeleteByExternalIDHandler,
) *UserWebhookHandler {
	return &UserWebhookHandler{
		getUserByExternalIDHandler:    getUserByExternalIDHandler,
		createUserHandler:             createUserHandler,
		updateUserHandler:             updateUserHandler,
		deleteUserByExternalIDHandler: deleteUserByExternalIDHandler,
	}
}

func (h *UserWebhookHandler) Execute(c fiber.Ctx) error {
	clerkEvent := c.Locals("payload").(dto.ClerkEvent)

	switch clerkEvent.Type {
	case "user.created":
		var data dto.ClerkUserCreated
		if err := json.Unmarshal(clerkEvent.Data, &data); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Invalid request body",
			})
		}
		if err := validation.Struct(c, &data); err != nil {
			return err
		}
		if err := h.createUser(c, data); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "Failed to create user",
			})
		}
		return c.SendStatus(fiber.StatusCreated)

	case "user.updated":
		var data dto.ClerkUserUpdated
		if err := json.Unmarshal(clerkEvent.Data, &data); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Invalid request body",
			})
		}
		if err := validation.Struct(c, &data); err != nil {
			return err
		}
		if err := h.updateUser(c, data); err != nil {
			if errors.Is(err, errUserNotFound) {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"message": "User not found",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "Failed to update user",
			})
		}
		return c.SendStatus(fiber.StatusNoContent)

	case "user.deleted":
		var data dto.ClerkUserDeleted
		if err := json.Unmarshal(clerkEvent.Data, &data); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Invalid request body",
			})
		}
		if err := validation.Struct(c, &data); err != nil {
			return err
		}
		if err := h.deleteUser(c, data); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "Failed to delete user",
			})
		}
		return c.SendStatus(fiber.StatusNoContent)

	default:
		return c.SendStatus(fiber.StatusOK)
	}
}

func (h *UserWebhookHandler) createUser(c fiber.Ctx, data dto.ClerkUserCreated) error {
	existing, err := h.getUserByExternalIDHandler.Handle(c.Context(), data.ID)
	if err != nil {
		log.Printf("Error finding user by Clerk ID %s: %v", data.ID, err)
		return err
	}
	if existing != nil {
		return nil
	}

	_, err = h.createUserHandler.Handle(c.Context(), usercmd.CreateUserCommand{
		ClerkID:   data.ID,
		FirstName: data.FirstName,
		LastName:  data.LastName,
		Email:     data.Email,
		Banned:    *data.Banned,
	})
	if err != nil {
		log.Printf("Error creating user with Clerk ID %s: %v", data.ID, err)
		return err
	}
	return nil
}

func (h *UserWebhookHandler) updateUser(c fiber.Ctx, data dto.ClerkUserUpdated) error {
	user, err := h.getUserByExternalIDHandler.Handle(c.Context(), data.ID)
	if err != nil {
		log.Printf("Error finding user by Clerk ID %s: %v", data.ID, err)
		return err
	}
	if user == nil {
		return errUserNotFound
	}

	if err := h.updateUserHandler.Handle(c.Context(), usercmd.UpdateUserCommand{
		User:      user,
		FirstName: data.FirstName,
		LastName:  data.LastName,
		Email:     data.Email,
		Banned:    *data.Banned,
	}); err != nil {
		log.Printf("Error updating user with Clerk ID %s: %v", data.ID, err)
		return err
	}
	return nil
}

func (h *UserWebhookHandler) deleteUser(c fiber.Ctx, data dto.ClerkUserDeleted) error {
	if err := h.deleteUserByExternalIDHandler.Handle(c.Context(), data.ID); err != nil {
		log.Printf("Error deleting user with Clerk ID %s: %v", data.ID, err)
		return err
	}
	return nil
}
