package middleware

import (
	"ares/config"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"net/http"
)

func ValidateToken(encodedToken string) (*jwt.Token, error) {
	conf := config.Get()
	secret := conf.Security.JWT

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

		tokenString := authHeader[len(BearerSchema):]
		token, err := ValidateToken(tokenString)
		if err != nil || !token.Valid {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		claims := token.Claims.(jwt.MapClaims)
		id := claims["accountId"].(string)

		ctx.Set("accountId", id)
		ctx.Next()
	}
}
