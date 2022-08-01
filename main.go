package main

import (
	"github.com/labstack/echo/v4"
	"time"

	"net/http"
)

const Host = "auth.almeng.kr"

const Day = time.Hour * 24
const Week = Day * 7
const Month = Week * 30
const Year = 31556952 * time.Second

//const API-GATEWAY = "api/"
func foo(c echo.Context) error {
	return c.JSON(http.StatusOK, "Hello, World!")
}

func main() {

	tok := User{
		Class:    "admin",
		ID:       "62e76e88c6100e00e162def9",
		Email:    "admin",
		Password: "ab32652fbb482c5800715de78cce98cef10dd719a21236a5ae5fd53b6aee0257",
	}.Claim().signRefreshToken()
	println("admin Token: " + tok)

	e := echo.New()
	//	With( Response { success, message, error, data }).routes(
	// returns token
	//	[POST] api/auth/login => require(body:{ username | email , password})
	e.POST("/auth/login", login)

	// auth with token and return user
	//	[GET] api/auth/me => require(header: { x-access-token})
	e.GET("/auth", auth)

	// return new token using existing token
	//	[GET] api/auth/refresh => require(header: { x-access-token})
	e.GET("/auth/refresh", refresh)
	//)

	//	With( Response { success, message, error, data } , Header{x-access-token}).routes(
	//
	// 	return user lists
	//	[GET] api/users
	e.GET("/users", indexUser)

	// 	get user of {username}
	//	[GET] api/users/{username}
	e.GET("/users/:username", getUser)

	//	create user
	//	[POST] api/users => require(body: { username, password, email })
	e.POST("/users", createUser)

	//	update user
	//	[PUT] api/users/{username} => require(body: {})
	e.PUT("/users/:username", updateUser)

	//	delete user
	//	[DELETE] api/users/{username}
	e.DELETE("/users/:username", deleteUser)
	//	)

	// 	send user a password reset link with token
	//	[PATCH] api/users/password => require(body: { email }), response( code: 202 )
	//	// update password verifying with token
	//	[PATCH] api/users/password => require(body: { token, newPassword }), response(code:204)
	//	// update password using old one
	//	[PATCH] api/users/password => require(header: {x-access-token}, body: {old, new}), response(code:204)
	e.PATCH("/users/password", foo)
	e.Logger.Fatal(e.Start(":1323"))
}
