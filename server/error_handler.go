package server

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ParamMissingError(ctx *gin.Context, n string) {
	ctx.AbortWithError(http.StatusBadRequest,
		fmt.Errorf("no %s found in request", n))
}

func InvalidSignature(ctx *gin.Context) {
	ctx.AbortWithError(http.StatusBadRequest,
		errors.New("invalid signature"))
}
