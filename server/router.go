package server

import (
	"fmt"
	"tokenguard/configuration"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type Router struct {
	conf   *configuration.MainConfiguration
	server *gin.Engine
}

func NewRouter(cfg *configuration.MainConfiguration) *Router {

	app := gin.New()

	app.Use(errorHandler)

	app.Any("/*proxy", func(ctx *gin.Context) { PostHandler(ctx, cfg.Entities, cfg.Server.Proxy) })

	return &Router{server: app, conf: cfg}
}

func (r *Router) Start() error {
	log.Warn().Msg("Router:: server start")
	return r.server.Run(fmt.Sprintf(":%d", r.conf.Server.Port))

}

func errorHandler(c *gin.Context) {
	c.Next()

	for _, err := range c.Errors {
		c.JSON(c.Copy().Writer.Status(), err)
		return
	}

}
