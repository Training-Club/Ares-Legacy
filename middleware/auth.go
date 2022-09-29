package middleware

import (
	"ares/config"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"net/http"
	"strconv"
)

// ValidateToken validates the provided encoded token against the
// provided public key.
func ValidateToken(encodedToken string, publicKey string) (*jwt.Token, error) {
	return jwt.Parse(encodedToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("invalid token %v", token.Header["alg"])
		}

		return []byte(publicKey), nil
	})
}

func ValidateRequest() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		const BearerSchema = "Bearer "

		conf := config.Get()
		accessTokenPublicKey := conf.Auth.AccessTokenPublicKey
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

		token, err := ValidateToken(tokenString, accessTokenPublicKey)
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
