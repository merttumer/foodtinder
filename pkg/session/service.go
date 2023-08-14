package session

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrExpiredSession = errors.New("session has expired")

type Repository interface {
	GetSession(context.Context, string) (UserSession, error)
	StoreSession(context.Context, string, time.Time) (UserSession, error)
}

type Service interface {
	Validate(context.Context, string) error

	create(context.Context) (UserSession, error)
}

type service struct {
	repo Repository
}

func (s *service) create(ctx context.Context) (UserSession, error) {
	sessionId := uuid.New().String()

	us, err := s.repo.StoreSession(ctx, sessionId, time.Now().Add(time.Hour))

	if err != nil {
		return UserSession{}, err
	}

	return us, nil
}

func (s *service) Validate(ctx context.Context, id string) error {
	us, err := s.repo.GetSession(ctx, id)
	if err != nil {
		return err
	}

	if time.Since(time.Unix(us.ExpireAt, 0)) > 0 {
		return ErrExpiredSession
	}

	return nil
}

func NewService(repo Repository) Service {
	return &service{repo}
}
