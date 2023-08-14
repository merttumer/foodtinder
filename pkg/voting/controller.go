package voting

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/merttumer/foodtinder/pkg/validation"
)

type Controller interface {
	RegisterRoutes(router fiber.Router)
}

type controller struct {
	svc Service
}

func NewController(svc Service) Controller {
	return &controller{svc}
}

func (h *controller) RegisterRoutes(router fiber.Router) {
	router.Post("/v1/votes", h.upsertVote)
	router.Get("/v1/votes", h.getVotes)
	router.Get("/v1/votes/:product_id", h.getAvgProductVotes)
}

// CreateSession godoc
// @Summary Returns average votes for a product
// @Description Returns average votes for a product
// @Tags root
// @Accept */*
// @Produce json
// @Success 200 {object} AvgVoteResponse
// @Param product_id path string true "Product ID"
// @Router /api/v1/votes/{product_id} [get]
func (h *controller) getAvgProductVotes(c *fiber.Ctx) error {
	productId := c.Params("product_id")

	if productId == "" {
		return fiber.ErrBadRequest
	}

	avg, err := h.svc.GetAvgProductVotes(c.Context(), productId)

	if err != nil {
		return fiber.ErrNotFound
	}

	return c.JSON(avg)
}

// GetVotes godoc
// @Summary Gets all votes given by a sessionid
// @Description Gets all votes given by a sessionid
// @Tags root
// @Accept */*
// @Produce json
// @Success 200 {object} []Vote
// @Router /api/v1/votes [get]
func (h *controller) getVotes(c *fiber.Ctx) error {
	sessionId := c.Cookies("session_id")

	if sessionId == "" {
		return fiber.ErrUnauthorized
	}

	votes, err := h.svc.GetVotes(c.Context(), sessionId)

	if votes == nil {
		votes = []Vote{}
	}

	if err != nil {
		return err
	}

	return c.JSON(votes)
}

// UpsertVotes godoc
// @Summary Inserts or updates a given vote for a product
// @Description Inserts or updates a given vote for a product
// @Tags root
// @Accept application/json
// @Produce json
// @Param body body VoteRequest true "Vote"
// @Success 200 {object} Vote
// @Router /api/v1/votes [post]
func (h *controller) upsertVote(c *fiber.Ctx) error {
	sessionId := c.Cookies("session_id")

	if sessionId == "" {
		fmt.Println("session id is empty", c.Context().RemoteAddr())
		return fiber.ErrUnauthorized
	}

	voteRequest := VoteRequest{}
	err := c.BodyParser(&voteRequest)

	validationErrors := validation.ValidateVoteRequest(voteRequest)
	if len(validationErrors) > 0 {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(validationErrors)
	}

	if err != nil {
		fmt.Println("error:", err.Error())
		return fiber.ErrBadRequest
	}

	vote, err := h.svc.Vote(c.Context(), sessionId, voteRequest.ProductID, voteRequest.Score)

	if err != nil {
		return err
	}

	c.JSON(vote)

	return nil
}
