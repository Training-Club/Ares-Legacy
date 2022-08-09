package util

import (
	"ares/config"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func ParseToken(tokenString string) (primitive.ObjectID, error) {
	var objectId primitive.ObjectID

	conf := config.Get()
	secret := []byte(conf.Auth.JWT)

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return secret, nil
	})

	if err != nil {
		return objectId, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		idString := claims["accountId"].(string)
		objectId, objectIdErr := primitive.ObjectIDFromHex(idString)
		return objectId, objectIdErr
	}

	return objectId, fmt.Errorf("bad token on %v", tokenString)
}

func GenerateToken(accountId string) (string, error) {
	conf := config.Get()
	secret := []byte(conf.Auth.JWT)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"accountId": accountId,
	})

	tokenString, err := token.SignedString(secret)
	return tokenString, err
}
