package mongodbstore

import (
	"context"
	"fmt"
	"time"

	envvars "github.com/merttumer/foodtinder/pkg/config/env-vars"
	"github.com/merttumer/foodtinder/pkg/session"
	"github.com/merttumer/foodtinder/pkg/voting"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

const (
	sessionCollection = "sessions"
	votesCollection   = "votes"
)

type Repository interface {
	GetSession(ctx context.Context, sessionId string) (session.UserSession, error)
	StoreSession(ctx context.Context, sessionId string, expire time.Time) (session.UserSession, error)

	UpsertVote(ctx context.Context, sessionId string, productId string, score int) (voting.Vote, error)
	GetSessionVotes(ctx context.Context, sessionId string) ([]voting.Vote, error)
	GetAvgProductVotes(ctx context.Context, productId string) (voting.AvgVoteResponse, error)
}

type SessionData struct {
	SessionID string
	ExpireAt  primitive.DateTime
}

type AvgVoteData struct {
	Avg       float64 `bson:"avg"`
	VoteCount int     `bson:"votecount"`
	ProductID string  `bson:"productid"`
}

type VoteData struct {
	SessionID string `bson:"sessionid"`
	ProductID string `bson:"productid"`
	Score     int    `bson:"score"`
}

type store struct {
	connectTimeout time.Duration
	uri            string
	client         *mongo.Client
	db             *mongo.Database
	pingTimeout    time.Duration
	database       string
}

func NewMongoStore(env envvars.Mongo) (Repository, error) {
	s := &store{
		connectTimeout: 5 * time.Second,
		uri:            env.URI,
		pingTimeout:    env.PingTimeout,
		database:       env.Database,
	}

	cctx, ccf := context.WithTimeout(context.Background(), s.connectTimeout)
	defer ccf()
	opts := options.Client()
	opts.ApplyURI(s.uri)
	c, err := mongo.Connect(cctx, opts)

	if err != nil {
		return nil, fmt.Errorf("cannot connect to mongodb, %s", err.Error())
	}

	s.client = c
	pctx, pcf := context.WithTimeout(context.Background(), s.pingTimeout)
	defer pcf()

	if err := s.client.Ping(pctx, readpref.Primary()); err != nil {
		return nil, fmt.Errorf("pinging failed, %s", err.Error())
	}

	s.db = c.Database(s.database)

	return s, nil
}

// GetSession implements Repository.
func (s *store) GetSession(ctx context.Context, id string) (session.UserSession, error) {
	c := s.db.Collection(sessionCollection)

	var sd SessionData

	filter := bson.M{"sessionid": id}
	err := c.FindOne(ctx, filter).Decode(&sd)

	if err != nil {
		return session.UserSession{}, fmt.Errorf("cannot find session, %s", err.Error())
	}

	return session.UserSession{
		SessionID: sd.SessionID,
		ExpireAt:  sd.ExpireAt.Time().Unix(),
	}, nil
}

// StoreSession implements Repository.
func (s *store) StoreSession(ctx context.Context, id string, expire time.Time) (session.UserSession, error) {
	c := s.db.Collection(sessionCollection)

	_, err := c.InsertOne(ctx, SessionData{
		SessionID: id,
		ExpireAt:  primitive.NewDateTimeFromTime(expire),
	})

	if err != nil {
		return session.UserSession{}, fmt.Errorf("cannot insert session, %s", err.Error())
	}

	return session.UserSession{
		SessionID: id,
		ExpireAt:  expire.Unix(),
	}, nil
}

// GetAvgProductVotes implements Repository.
func (s *store) GetAvgProductVotes(ctx context.Context, productId string) (voting.AvgVoteResponse, error) {
	c := s.db.Collection(votesCollection)

	matchStage := bson.D{{Key: "$match", Value: bson.D{{Key: "productid", Value: productId}}}}

	groupStage := bson.D{
		{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$productid"},
			{Key: "avg_score", Value: bson.D{{Key: "$avg", Value: "$score"}}},
			{Key: "vote_count", Value: bson.D{{Key: "$sum", Value: 1}}},
		}},
	}

	projectStage := bson.D{
		{Key: "$project", Value: bson.D{
			{Key: "avg", Value: "$avg_score"},
			{Key: "votecount", Value: "$vote_count"},
			{Key: "productid", Value: "$_id"},
		}},
	}

	cursor, err := c.Aggregate(ctx, mongo.Pipeline{matchStage, groupStage, projectStage})
	if err != nil {
		return voting.AvgVoteResponse{}, err
	}
	defer cursor.Close(ctx)

	var result AvgVoteData
	if cursor.Next(ctx) {
		err := cursor.Decode(&result)
		if err != nil {
			return voting.AvgVoteResponse{}, err
		}
	}

	if result.ProductID == "" {
		return voting.AvgVoteResponse{}, fmt.Errorf("no votes found for product %s", productId)
	}

	return voting.AvgVoteResponse{
		Avg:       result.Avg,
		VoteCount: result.VoteCount,
		ProductID: result.ProductID,
	}, nil
}

// GetSessionVotes implements Repository.
func (s *store) GetSessionVotes(ctx context.Context, sessionId string) ([]voting.Vote, error) {
	c := s.db.Collection(votesCollection)

	filter := bson.D{{Key: "sessionid", Value: sessionId}}

	cursor, err := c.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var votes []voting.Vote
	for cursor.Next(ctx) {
		var vote VoteData
		err := cursor.Decode(&vote)
		if err != nil {
			return nil, err
		}
		votes = append(votes, voting.Vote{
			ProductID: vote.ProductID,
			SessionID: vote.SessionID,
			Score:     vote.Score,
		})
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return votes, nil
}

// UpsertVote implements Repository.
func (s *store) UpsertVote(ctx context.Context, sessionId string, productId string, score int) (voting.Vote, error) {
	c := s.db.Collection(votesCollection)

	filter := bson.M{
		"sessionid": sessionId,
		"productid": productId,
	}

	update := bson.M{
		"$set": bson.M{
			"score": score,
		},
	}

	opts := options.Update().SetUpsert(true)

	_, err := c.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return voting.Vote{}, err
	}

	return voting.Vote{
		ProductID: productId,
		SessionID: sessionId,
		Score:     score,
	}, nil
}
