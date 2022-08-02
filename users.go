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
	service := "Index User"
	defer println()
	LogContext(c, service+":")
	if _authRequest(c, true) {
		return UnauthorizedRequest(c, service)
	}
	response := new(Response)
	users, err := DBFindUsers(nil)
	if err != nil {
		return RequestFailed(c, service, http.StatusServiceUnavailable, "Error occurred during indexing users", response)
	}

	response.Success = true
	response.Data = struct {
		Users []User `json:"users"`
	}{users.Slice()}
	return c.JSON(http.StatusOK, response)
}

func RequestFailed(c echo.Context, service string, status int, error string, response *Response) error {
	response.Error = error
	LogContext(c, service+":", "Request Failed:", response.Error, "Status Code:", status)
	return c.JSON(status, response)
}

// 	get user of {username}
//	[GET] api/users/{username}
func getUser(c echo.Context) error {
	service := "Get User"
	defer println()

	if !_authRequest(c, false) {
		return UnauthorizedRequest(c, service)
	}
	name := c.Param("username")
	LogContext(c, service+": "+name)
	token := c.Request().Header.Get(TokenHeader)
	claim, _ := verifyToken(token)
	if claim.Class != ClassAdmin && claim.Email != name {
		return UnauthorizedRequest(c, service)
	}
	filter := bson.D{{"email", name}}
	user, err := DBFindUserOne(filter)
	if err != nil {
		return handleDBError(c, service, err)
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
	service := "Create User"
	defer println()
	LogContext(c, service+":")
	request := new(UserRequest)
	err := c.Bind(request)
	response := new(Response)
	response.Success = false
	if err != nil {
		return RequestFailed(c, service, http.StatusBadRequest, "Invalid Request", response)
	}

	if !isValidEmail(request.Email) {
		return RequestFailed(c, service, http.StatusBadRequest, "Invalid Email Format", response)
	}

	if !checkPasswordFormat(request.Password) {
		return RequestFailed(c, service, http.StatusBadRequest, "Invalid Password Format", response)
	}
	user, err := DBCreateUser(*(request.NewUser(ClassDefault)))
	if err != nil {
		return handleDBError(c, service, err)
	}
	response.Success = true
	at := user.Claim().signAccessToken()
	rt := user.Claim().signRefreshToken()
	response.Data = struct {
		User   User   `json:"user"`
		AToken string `json:"access_token"`
		RToken string `json:"refresh_token"`
	}{*user, at, rt}
	return c.JSON(http.StatusOK, response)
}

//	update user
//	[PUT] api/users/{username} => require(body: user)
func updateUser(c echo.Context) error {
	service := "Update User"
	defer println()
	if !_authRequest(c, false) {
		return UnauthorizedRequest(c, service)
	}
	request := new(UserRequest)
	name := c.Param("username")
	LogContext(c, service+": "+name)
	token := c.Request().Header.Get(TokenHeader)
	claim, _ := verifyToken(token)
	if claim.Class != ClassAdmin && claim.Email != name {
		return UnauthorizedRequest(c, service)
	}
	err := c.Bind(request)
	response := new(Response)
	response.Success = false
	if err != nil {
		return RequestFailed(c, service, http.StatusBadRequest, "Invalid Request", response)
	}
	filter := bson.D{{"email", name}}
	err = DBUpdateUser(filter, *request.NewUser(claim.Class))
	if err != nil {
		return handleDBError(c, service, err)
	}
	response.Success = true
	return c.JSON(http.StatusOK, response)
}

//	delete user
//	[DELETE] api/users/{username}
func deleteUser(c echo.Context) error {
	service := "Delete User"
	defer println()
	if !_authRequest(c, false) {
		return UnauthorizedRequest(c, service)
	}
	name := c.Param("username")
	LogContext(c, service+": "+name)
	res := new(Response)
	res.Success = false
	token := c.Request().Header.Get(TokenHeader)
	claim, _ := verifyToken(token)
	if claim.Class != ClassAdmin && claim.Email != name {
		return UnauthorizedRequest(c, service)
	}
	err := DBDeleteUser(name)
	if err != nil {
		return handleDBError(c, service, err)
	}
	res.Success = true
	return c.JSON(http.StatusOK, res)
}
