package middleware

import (
	"ares/config"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"net/http"
	"strconv"
)

func ValidateToken(encodedToken string) (*jwt.Token, error) {
	conf := config.Get()
	secret := conf.Auth.JWT

	return jwt.Parse(encodedToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("invalid token %v", token.Header["alg"])
		}

		return []byte(secret), nil
	})
}

func ValidateRequest() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		const BearerSchema = "Bearer "
		authHeader := ctx.GetHeader("Authorization")

		if len(authHeader) < 7 {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "bad authorization header"})
			return
		}

		var tokenString string
		var err error
		tokenString = authHeader[len(BearerSchema):]

		// if the request is sent with a prefixed double-quote we need
		// to unquote the token before attempting to verify it
		if string(tokenString[0]) == `"` {
			tokenString, err = strconv.Unquote(tokenString)
		}

		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "failed to unquote token"})
			return
		}

		token, err := ValidateToken(tokenString)
		if err != nil || !token.Valid {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "token invalid: " + err.Error()})
			return
		}

		claims := token.Claims.(jwt.MapClaims)
		id := claims["accountId"].(string)

		ctx.Set("accountId", id)
		ctx.Next()
	}
}
