package server

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"
	"time"

	"net/http"
	"os"
	"tokenguard/configuration"

	"github.com/rs/zerolog/log"

	"github.com/gin-gonic/gin"
)

type Router struct {
	conf   *configuration.MainConfiguration
	server *gin.Engine
	// shutdownChannel chan os.Signal
	// ready           bool
}

func NewRouter(cfg *configuration.MainConfiguration) *Router {

	gin.SetMode(cfg.Ginmode)

	app := gin.Default()

	return &Router{server: app, conf: cfg}
}

func (r *Router) Start() error {

	r.server.Use(errorHandler)

	r.server.Any("/*proxy",
		func(ctx *gin.Context) {
			Handler(ctx, r.conf.Entities, r.conf.Server.Proxy)
		})

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", r.conf.Server.Port),
		Handler: r.server.Handler(),
	}

	go func() {

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Msgf("error staring server: %s", err.Error())
			os.Exit(2)
		}
	}()

	quit := make(chan os.Signal, 1)

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal().Msgf("Server is being forcefully shut down: %s", err.Error())
		os.Exit(2)
	}

	<-ctx.Done()

	log.Warn().Msgf("Server exiting")

	return nil
}

func errorHandler(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Next()

	for _, err := range c.Errors {
		c.AbortWithStatusJSON(c.Copy().Writer.Status(), err)
		return
	}

}
