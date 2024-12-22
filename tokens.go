package main

import (
	"crypto/hmac"
	"crypto/sha512"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TokensUtils struct {
	secret []byte
}

type AccessTokenClaims struct {
	Ip     string `json:"ip"`
	UserId string `json:"user_id"`
	jwt.RegisteredClaims
}

type RefreshTokenClaims struct {
	AccessToken string
	AccessTokenClaims
}

// генерирует пару токенов
func (tu *TokensUtils) New(ip string, userId string) (string, []byte, error) {
	claims := AccessTokenClaims{
		ip,
		userId,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(3 * time.Hour)),
		},
	}
	log.Println(claims)

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	accessStr, err := accessToken.SignedString(tu.secret)
	if err != nil {
		panic(err)
	}

	refreshToken := tu.GenRefreshToken(accessStr)

	return accessStr, refreshToken, nil
}

// генерирует refresh токен из access токена
func (tu *TokensUtils) GenRefreshToken(acc string) []byte {
	// TODO: возможно hmac не нужно пересоздавать каждый раз
	h := hmac.New(sha512.New, tu.secret)
	h.Write([]byte(acc))
	refreshToken := h.Sum(nil)

	return refreshToken
}

// проверяет пару токенов на связанность
func (tu *TokensUtils) ValTokensPair(acc string, ref string) bool {
	validRef := tu.GenRefreshToken(acc)

	return hmac.Equal(validRef, []byte(ref))
}

var ErrUnknownClaimsType = errors.New("bad access token payload")

func (tu *TokensUtils) ParseAccessToken(acc string) (*AccessTokenClaims, error) {
	// разбор access токена с валидацией alg
	token, err := jwt.ParseWithClaims(acc, &AccessTokenClaims{}, func(t *jwt.Token) (interface{}, error) {
		// если alg не HMAC
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", t.Header["alg"])
		}

		return tu.secret, nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*AccessTokenClaims); ok {
		return claims, nil
	}

	return nil, ErrUnknownClaimsType
}
