package voting

import (
	"context"
	"errors"
	"fmt"

	"github.com/merttumer/foodtinder/pkg/session"
)

var (
	ErrInvalidSession = errors.New("invalid session")
)

type Repository interface {
	UpsertVote(ctx context.Context, sessionId string, productId string, score int) (Vote, error)
	GetSessionVotes(ctx context.Context, sessionId string) ([]Vote, error)
	GetAvgProductVotes(ctx context.Context, productId string) (AvgVoteResponse, error)
}

type Service interface {
	Vote(ctx context.Context, sessionId string, productId string, score int) (Vote, error)
	GetVotes(ctx context.Context, sessionId string) ([]Vote, error)
	GetAvgProductVotes(ctx context.Context, productId string) (AvgVoteResponse, error)
}

type service struct {
	repo       Repository
	sessionSvc session.Service
}

func NewService(repo Repository, sessionSvc session.Service) Service {
	return &service{repo, sessionSvc}
}

// GetAvgProductVotes implements Service.
func (s *service) GetAvgProductVotes(ctx context.Context, productId string) (AvgVoteResponse, error) {
	return s.repo.GetAvgProductVotes(ctx, productId)
}

// GetVotes implements Service.
func (s *service) GetVotes(ctx context.Context, sessionId string) ([]Vote, error) {
	err := s.sessionSvc.Validate(ctx, sessionId)
	if err != nil {
		fmt.Println("err validating session", err.Error())
		return nil, ErrInvalidSession
	}
	return s.repo.GetSessionVotes(ctx, sessionId)
}

// Vote implements Service.
func (s *service) Vote(ctx context.Context, sessionId string, productId string, score int) (Vote, error) {
	err := s.sessionSvc.Validate(ctx, sessionId)
	if err != nil {
		return Vote{}, ErrInvalidSession
	}
	return s.repo.UpsertVote(ctx, sessionId, productId, score)
}
