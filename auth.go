package main

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"net/http"
)

func handleDBError(c echo.Context, err error) error {
	var status int
	response := Response{Success: false}
	switch err {
	case UserNotExistError:
		response.Error = UserNotExistError.Error()
		status = http.StatusNotFound
	case UserAlreadyExistError:
		response.Error = UserAlreadyExistError.Error()
		status = http.StatusConflict
	default:
		print(err.Error())
		response.Error = "Error occurred during indexing: " + err.Error()
		status = http.StatusServiceUnavailable
	}
	return c.JSON(status, response)
}

// returns token
//	[POST] api/auth/login => require(body:{ username | email , password})
func login(c echo.Context) error {
	defer println()
	println("login request")
	request := new(UserRequest)
	err := c.Bind(request)
	response := new(Response)
	response.Success = false

	if err != nil {
		print(err)
		response.Error = "Invalid Login Request"
		return c.JSON(http.StatusBadRequest, response)
	}

	fmt.Println(request.Email, "\n", request.Password)
	filter := bson.D{{"email", request.Email}}
	user, err := DBFindUserOne(filter)

	if err != nil {
		return handleDBError(c, err)
	}

	if !checkPasswordFormat(request.Password) {
		response.Error = "Invalid Password format"
		return c.JSON(http.StatusBadRequest, response)
	}

	if user.Password != request.Password {
		response.Error = "Password Incorrect"
		return c.JSON(http.StatusUnauthorized, response)
	}
	AccessToken := user.Claim().signAccessToken()
	RefreshToken := user.Claim().signRefreshToken()
	response.Success = true
	response.Data = struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}{AccessToken, RefreshToken}
	return c.JSON(http.StatusOK, response)
}

// auth with token and return user
// used by other api gateways
//	[GET] api/auth/me => require(header: { x-access-token})
func auth(c echo.Context) error {
	token := c.Request().Header.Get("x-access-token")
	claim, ok := verifyToken(token)
	if !ok {
		return UnauthorizedRequest(c)
	}
	response := new(Response)
	response.Success = true
	response.Data = claim
	return c.JSON(http.StatusOK, response)
}

// return new token using existing token
//	[GET] api/auth/refresh => require(header: { x-access-token })
func refresh(c echo.Context) error {
	request := c.Request()
	token := request.Header.Get("x-access-token")
	claim, ok := verifyToken(token)
	if !ok {
		return UnauthorizedRequest(c)
	}

	tok := claim.signAccessToken()
	response := new(Response)
	response.Success = true
	response.Data = struct {
		Token string `json:"access_token"`
	}{tok}

	return c.JSON(http.StatusOK, response)
}
