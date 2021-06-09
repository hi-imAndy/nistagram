package util

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"net/http"
	"strings"
	"time"
)

const ExpiresIn = 86400
const TokenSecret = "token_secret"

type TokenClaims struct {
	LoggedUserId uint `json:"loggedUserId"`
	jwt.StandardClaims
}

func CreateToken(userId uint, issuer string) (string, error) {
	claims := TokenClaims{LoggedUserId: userId, StandardClaims: jwt.StandardClaims{
		ExpiresAt: time.Now().Unix() + ExpiresIn,
		IssuedAt:  time.Now().Unix(),
		Issuer:    issuer,
	}}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(TokenSecret))
}

func GetToken(header http.Header) (string, error) {
	reqToken := header.Get("Authorization")
	splitToken := strings.Split(reqToken, "Bearer ")
	if len(splitToken) != 2 {
		return "", fmt.Errorf("NO_TOKEN")
	}
	return strings.TrimSpace(splitToken[1]), nil
}

func GetLoggedUserIDFromToken(r *http.Request) uint {
	tokenString, err := GetToken(r.Header)
	if err != nil {
		fmt.Println(err)
		return 0
	}
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(TokenSecret), nil
	})
	if err != nil {
		fmt.Println(err)
		return 0
	}
	if claims, ok := token.Claims.(*TokenClaims); ok && token.Valid {
		return claims.LoggedUserId
	} else {
		fmt.Println(err)
		return 0
	}
}
