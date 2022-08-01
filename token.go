package main

import (
	"errors"
	"github.com/golang-jwt/jwt"
	_ "github.com/golang-jwt/jwt"
	"time"
)

type TokenClaim struct {
	Email string `json:"email"`
	ID    string `json:"id"`
	Class string `json:"class"`
	jwt.StandardClaims
}

var Method = jwt.SigningMethodHS256
var PrivateKey = "a3IuYWxtZW5nLmF1dGgucHJpdmF0ZS5rZXkuMjAyMjA3MzE"

func key(token *jwt.Token) (interface{}, error) {
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		ErrUnexpectedSigningMethod := errors.New("unexpected signing method")
		return nil, ErrUnexpectedSigningMethod
	}
	return []byte(PrivateKey), nil
}

func StandardAccessClaim() jwt.StandardClaims {
	now := time.Now()
	return jwt.StandardClaims{
		IssuedAt:  now.Unix(),
		ExpiresAt: now.Add(time.Minute * 20).Unix(),
		Issuer:    Host,
	}
}

func StandardRefreshClaim() jwt.StandardClaims {
	now := time.Now()
	return jwt.StandardClaims{
		IssuedAt:  now.Unix(),
		ExpiresAt: now.Add(Week).Unix(),
		Issuer:    Host,
	}
}

func (claim TokenClaim) signAccessToken() string {
	return parseClaim(claim, StandardAccessClaim())
}

func (claim TokenClaim) signRefreshToken() string {
	return parseClaim(claim, StandardRefreshClaim())
}

func parseClaim(claim TokenClaim, stdClaim jwt.StandardClaims) string {
	claim.StandardClaims = stdClaim
	atoken := jwt.NewWithClaims(jwt.SigningMethodHS256, &claim)
	signedAuthToken, err := atoken.SignedString([]byte(PrivateKey))

	if err != nil {
		print("error occurred during signing token: ")
		println(err)
	}
	println(signedAuthToken)
	_, is := verifyToken(signedAuthToken)
	println("Is token Valid: ", is)
	return signedAuthToken
}

func verifyToken(token string) (*TokenClaim, bool) {
	claim := &TokenClaim{}
	tok, err := jwt.ParseWithClaims(token, claim, key)
	if err != nil {
		return nil, false
	}
	return claim, tok.Valid
}
