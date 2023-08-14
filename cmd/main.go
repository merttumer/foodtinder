package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"

	swagger "github.com/arsmn/fiber-swagger/v2"
	"github.com/go-kit/log"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	_ "github.com/merttumer/foodtinder/docs/swaggerdocs"
	envvars "github.com/merttumer/foodtinder/pkg/config/env-vars"
	"github.com/merttumer/foodtinder/pkg/session"
	mongodbstore "github.com/merttumer/foodtinder/pkg/store/mongo"
	"github.com/merttumer/foodtinder/pkg/voting"
)

// @title FoodTinder Swagger API
// @version 2.0
// @description This is a sample server server.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /
// @schemes http
func main() {
	var l log.Logger
	{
		l = log.NewLogfmtLogger(os.Stdout)
		l = log.With(l, "time", log.DefaultTimestampUTC)
	}
	envs, err := envvars.NewEnvvars()

	if err != nil {
		_ = l.Log("error", err)
		return
	}

	app := fiber.New(fiber.Config{
		// Override default error handler
		ErrorHandler: func(ctx *fiber.Ctx, err error) error {
			// Status code defaults to 500
			code := fiber.StatusInternalServerError

			// Retrieve the custom status code if it's a *fiber.Error
			var e *fiber.Error
			if errors.As(err, &e) {
				code = e.Code
			}

			return ctx.Status(code).JSON(map[string]interface{}{
				"error": err.Error(),
			})
		},
	})

	app.Use(recover.New(recover.Config{
		EnableStackTrace: true,
	}))

	app.Get("/swagger/*", swagger.HandlerDefault)

	mongoStore, err := mongodbstore.NewMongoStore(envs.Mongo)
	if err != nil {
		_ = l.Log("error", err)
		return
	}

	sessionSvc := session.NewService(mongoStore)
	votingSvc := voting.NewService(mongoStore, sessionSvc)

	sessionController := session.NewController(l, sessionSvc)
	votingController := voting.NewController(votingSvc)

	router := app.Group("/api")

	router.Get("/health", func(c *fiber.Ctx) error {
		c.Status(fiber.StatusOK)
		return nil
	})

	sessionController.RegisterRoutes(router)
	votingController.RegisterRoutes(router)

	errs := make(chan error)

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errs <- errors.New((<-c).String())
	}()

	go func() {
		errs <- app.Listen(":" + envs.Service.Port)
	}()

	err = <-errs

	_ = l.Log("error", err.Error())

	ctx, cf := context.WithTimeout(context.Background(), envs.Service.ShutdownTimeout)
	defer cf()
	if err := app.ShutdownWithContext(ctx); err != nil {
		_ = l.Log("error", err.Error())
	}
}
