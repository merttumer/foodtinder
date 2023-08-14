package session

import (
	"time"

	"github.com/go-kit/log"
	"github.com/gofiber/fiber/v2"
)

type Controller interface {
	RegisterRoutes(router fiber.Router)
}

type controller struct {
	svc Service
	l   log.Logger
}

func NewController(l log.Logger, svc Service) Controller {
	return &controller{svc, l}
}

func (h *controller) RegisterRoutes(router fiber.Router) {
	router.Post("/v1/session", h.createSessionHandler)
}

// CreateSession godoc
// @Summary Creates a new session
// @Description New session id is stored in a cookie with key "session_id"
// @Tags root
// @Accept */*
// @Produce json
// @Success 200
// @Router /api/v1/session [post]
func (h *controller) createSessionHandler(c *fiber.Ctx) error {
	newSession, err := h.svc.create(c.Context())

	if err != nil {
		h.l.Log("error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "session could not be created")
	}

	c.ClearCookie("session_id")
	c.Cookie(&fiber.Cookie{
		Name:    "session_id",
		Value:   newSession.SessionID,
		Expires: time.Unix(newSession.ExpireAt, 0),
		Secure:  true,
	})

	return nil
}
