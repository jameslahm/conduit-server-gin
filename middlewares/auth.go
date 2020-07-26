package middlewares

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jameslahm/conduit-server-gin/models"
)

// TokenInput toekn header
type TokenInput struct {
	TokenHeader string `header:"Authorization" binding:"required"`
}

// Authenticate auth
func Authenticate(c *gin.Context) (*models.JwtClaims, error) {

	var authHeader TokenInput
	if err := c.ShouldBindHeader(&authHeader); err != nil {
		return nil, err
	}

	splitStrings := strings.Split(authHeader.TokenHeader, " ")
	if len(splitStrings) != 2 {
		return nil, errors.New("error: Bad Request")
	}
	claims, err := models.VerifyToken(splitStrings[1])
	return claims, err

}
