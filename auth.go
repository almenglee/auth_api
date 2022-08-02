package main

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"net/http"
)

func handleDBError(c echo.Context, service string, err error) error {
	var status int
	response := Response{Success: false}
	var e string
	switch err {
	case UserNotExistError:
		e = UserNotExistError.Error()
		status = http.StatusNotFound

	case UserAlreadyExistError:
		e = UserAlreadyExistError.Error()
		status = http.StatusConflict
	default:
		e = "Error occurred during indexing: " + err.Error()
		status = http.StatusServiceUnavailable
	}
	return RequestFailed(c, service, status, e, &response)
}

// returns token
//	[POST] api/auth/login => require(body:{ username | email , password})
func login(c echo.Context) error {
	service := "Login"
	defer println()
	LogContext(c, service+":")
	request := new(UserRequest)
	err := c.Bind(request)
	response := new(Response)
	response.Success = false
	if err != nil {
		return RequestFailed(c, service, http.StatusBadRequest, "Invalid Login Request", response)
	}

	fmt.Println(request.Email, "\n", request.Password)
	filter := bson.D{{"email", request.Email}}
	user, err := DBFindUserOne(filter)

	if err != nil {
		return handleDBError(c, service, err)
	}

	if !checkPasswordFormat(request.Password) {
		return RequestFailed(c, service, http.StatusBadRequest, "Invalid Password format", response)
	}

	if user.Password != request.Password {
		return RequestFailed(c, service, http.StatusUnauthorized, "Password Incorrect", response)
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
	service := "Auth"
	defer println()
	LogContext(c, service+":")
	token := c.Request().Header.Get("x-access-token")
	claim, ok := verifyToken(token)
	if !ok {
		return UnauthorizedRequest(c, service)
	}
	response := new(Response)
	response.Success = true
	response.Data = claim
	return c.JSON(http.StatusOK, response)
}

// return new token using existing token
//	[GET] api/auth/refresh => require(header: { x-access-token })
func refresh(c echo.Context) error {
	service := "Refresh"
	defer println()
	LogContext(c, service+":")
	request := c.Request()
	token := request.Header.Get("x-access-token")
	claim, ok := verifyToken(token)
	if !ok {
		return UnauthorizedRequest(c, service)
	}

	tok := claim.signAccessToken()
	response := new(Response)
	response.Success = true
	response.Data = struct {
		Token string `json:"access_token"`
	}{tok}

	return c.JSON(http.StatusOK, response)
}
