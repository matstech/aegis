package server

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os/signal"
	"syscall"
	"time"

	"aegis/configuration"
	"net/http"
	"os"

	"github.com/rs/zerolog/log"

	"github.com/gin-gonic/gin"
)

type Router struct {
	conf   *configuration.MainConfiguration
	server *gin.Engine
}

func NewRouter(cfg *configuration.MainConfiguration) *Router {

	gin.SetMode(cfg.Ginmode)

	app := gin.Default()

	return &Router{server: app, conf: cfg}
}

func (r *Router) Start() error {

	r.server.Use(errorHandler)

	// probes port
	// probesSrv := &http.Server{
	// 	Addr: fmt.Sprintf(":%d", r.conf.Server.ProbesPort),
	// }

	// r.server.GET("/liveness", func(ctx *gin.Context) {
	// 	ctx.Status(http.StatusOK)
	// })
	// r.server.GET("/readiness", func(ctx *gin.Context) {
	// 	ctx.Status(http.StatusOK)
	// })

	r.server.Any("/*proxy",
		func(ctx *gin.Context) {
			r.Handler(ctx)
		})

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", r.conf.Server.Port),
		Handler: r.server.Handler(),
	}

	go func() {

		if r.conf.Server.Mode == "PLAIN" {

			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatal().Msgf("error starting server: %s", err.Error())
				os.Exit(2)
			}

		} else {
			//generic TLS policy
			srv.TLSConfig = &tls.Config{
				ClientAuth: tls.NoClientCert,
			}

			if r.conf.Server.Mode == "MTLS" {
				certPool, certPoolErr := buildCertPool(r.conf.Server.Tls.Cacert)

				if certPoolErr != nil {
					log.Fatal().Msgf("error loading certpool: %s", certPoolErr.Error())
					os.Exit(2)
				}

				srv.TLSConfig.ClientCAs = certPool
				srv.TLSConfig.ClientAuth = tls.RequireAndVerifyClientCert
			}

			if err := srv.ListenAndServeTLS(
				r.conf.Server.Tls.Certpath,
				r.conf.Server.Tls.Keypath); err != nil && err != http.ErrServerClosed {
				log.Fatal().Msgf("error starting server: %s", err.Error())
				os.Exit(2)
			}
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

func buildCertPool(cacertpath string) (*x509.CertPool, error) {
	certPool, _ := x509.SystemCertPool()

	if certPool == nil {
		certPool = x509.NewCertPool()
	}

	if cacertpath != "" {
		c, err := os.ReadFile(cacertpath)

		if err != nil {
			return nil, err
		}

		certPool.AppendCertsFromPEM(c)
	}

	return certPool, nil
}
