package envvars

import (
	"fmt"
	"time"

	"github.com/codingconcepts/env"
)

type Envvars struct {
	Service Service
	Mongo   Mongo
}

type Service struct {
	Port            string        `env:"SERVICE_PORT" default:"8080"`
	ShutdownTimeout time.Duration `env:"SERVICE_SHUTDOWN_TIMEOUT" default:"15s"`
}

type Mongo struct {
	URI         string        `env:"MONGO_URI" required:"true"`
	PingTimeout time.Duration `env:"MONGO_PING_TIMEOUT" default:"5s"`
	Database    string        `env:"MONGO_DATABASE" default:"foodtinder"`
}

func NewEnvvars() (*Envvars, error) {

	service := Service{}
	if err := env.Set(&service); err != nil {
		return nil, fmt.Errorf("cannot set env vars, %s", err.Error())
	}

	mongo := Mongo{}
	if err := env.Set(&mongo); err != nil {
		return nil, fmt.Errorf("cannot set env vars, %s", err.Error())
	}

	e := &Envvars{service, mongo}

	return e, nil
}
