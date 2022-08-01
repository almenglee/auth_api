package main

import (
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"net/http"
)

const TokenHeader = "x-access-token"

func _authRequest(c echo.Context, reqAdmin bool) bool {
	request := c.Request()
	token := request.Header.Get(TokenHeader)
	claim, ok := verifyToken(token)
	return ok && !(reqAdmin && claim.Class == "admin")
}

//With( Response { success, message, error, data } , Header{x-access-token})

// 	return user lists
//	[GET] api/users
func indexUser(c echo.Context) error {
	println("user index request")
	if _authRequest(c, true) {
		return UnauthorizedRequest(c)
	}
	response := new(Response)
	users, err := DBFindUsers(nil)
	if err != nil {
		response.Error = "Error occurred during indexing users"
		return c.JSON(http.StatusServiceUnavailable, response)
	}

	response.Success = true
	response.Data = struct {
		Users []User `json:"users"`
	}{users.Slice()}
	return c.JSON(http.StatusOK, response)
}

// 	get user of {username}
//	[GET] api/users/{username}
func getUser(c echo.Context) error {
	if !_authRequest(c, false) {
		return UnauthorizedRequest(c)
	}
	name := c.Param("username")
	token := c.Request().Header.Get(TokenHeader)
	claim, _ := verifyToken(token)
	if claim.Class != ClassAdmin && claim.Email != name {
		return UnauthorizedRequest(c)
	}
	filter := bson.D{{"email", name}}
	user, err := DBFindUserOne(filter)
	if err != nil {
		return handleDBError(c, err)
	}
	response := new(Response)
	response.Success = true
	response.Data = struct {
		User TokenClaim `json:"user"`
	}{*user.Claim()}
	return c.JSON(http.StatusOK, response)
}

//	create user
//	No header required
//	[POST] api/users => require(body: { username, password, email })
func createUser(c echo.Context) error {
	request := new(UserRequest)
	err := c.Bind(request)
	res := new(Response)
	res.Success = false
	if err != nil {
		res.Error = "Invalid Request"
		return c.JSON(http.StatusBadRequest, res)
	}

	if !isValidEmail(request.Email) {
		res.Error = "Invalid Email Format"
		return c.JSON(http.StatusBadRequest, res)
	}

	if !checkPasswordFormat(request.Password) {
		res.Error = "Invalid Password Format"
		return c.JSON(http.StatusBadRequest, res)
	}
	user, err := DBCreateUser(*(request.NewUser(ClassDefault)))
	if err != nil {
		return handleDBError(c, err)
	}
	res.Success = true
	at := user.Claim().signAccessToken()
	rt := user.Claim().signRefreshToken()
	res.Data = struct {
		User   User   `json:"user"`
		AToken string `json:"access_token"`
		RToken string `json:"refresh_token"`
	}{*user, at, rt}
	return c.JSON(http.StatusOK, res)
}

//	update user
//	[PUT] api/users/{username} => require(body: user)
func updateUser(c echo.Context) error {
	if !_authRequest(c, false) {
		return UnauthorizedRequest(c)
	}
	request := new(UserRequest)
	name := c.Param("username")
	token := c.Request().Header.Get(TokenHeader)
	claim, _ := verifyToken(token)
	if claim.Class != ClassAdmin && claim.Email != name {
		return UnauthorizedRequest(c)
	}
	err := c.Bind(request)
	res := new(Response)
	res.Success = false
	if err != nil {
		res.Error = "Invalid Request"
		return c.JSON(http.StatusBadRequest, res)
	}
	filter := bson.D{{"email", name}}
	err = DBUpdateUser(filter, *request.NewUser(claim.Class))
	if err != nil {
		return handleDBError(c, err)
	}
	res.Success = true
	return c.JSON(http.StatusOK, res)
}

//	delete user
//	[DELETE] api/users/{username}
func deleteUser(c echo.Context) error {
	if !_authRequest(c, false) {
		return UnauthorizedRequest(c)
	}
	name := c.Param("username")
	res := new(Response)
	res.Success = false
	token := c.Request().Header.Get(TokenHeader)
	claim, _ := verifyToken(token)
	if claim.Class != ClassAdmin && claim.Email != name {
		return UnauthorizedRequest(c)
	}
	err := DBDeleteUser(name)
	if err != nil {
		return handleDBError(c, err)
	}
	res.Success = true
	return c.JSON(http.StatusOK, res)
}
