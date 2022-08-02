package main

import (
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"io"
	"log"
	"net/http"
	"net/mail"
	"os"
	"reflect"
	"regexp"
)

type (
	UserRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	User struct {
		Class    string `bson:"class,omitempty"`
		ID       string `bson:"_id,omitempty"`
		Email    string `bson:"email,omitempty"`
		Password string `bson:"password,omitempty"`
	}
	//	With( Response { success, message, error, data })
	Response struct {
		Success bool        `json:"success"`
		Error   string      `json:"error"`
		Data    interface{} `json:"data"`
	}

	Class = string
)

var (
	ClassAdmin   Class = "admin"
	ClassUser    Class = "user"
	ClassDefault Class = ""
)

func (r *UserRequest) NewUser(class Class) *User {
	return &User{Class: class, Email: r.Email, Password: r.Password}
}

func (u *User) ToBson() (data bson.M) {
	var tagValue string

	data = bson.M{}
	element := reflect.ValueOf(u).Elem()

	for i := 0; i < element.NumField(); i += 1 {
		typeField := element.Type().Field(i)
		tag := typeField.Tag

		tagValue = tag.Get("bson")

		if tagValue == "-" {
			continue
		}

		switch element.Field(i).Kind() {
		case reflect.String:
			value := element.Field(i).String()
			data[tagValue] = value

		case reflect.Bool:
			value := element.Field(i).Bool()
			data[tagValue] = value

		case reflect.Int:
			value := element.Field(i).Int()
			data[tagValue] = value
		}
	}

	return
}

type Logger struct {
	logg *log.Logger
	f    *os.File
}

func Log(v ...any) {
	logger.logg.Println(v...)
}

func LogContext(c echo.Context, v ...any) {
	v = append(v, " FROM: "+c.RealIP())
	logger.logg.Println(v...)
}

func NewLogger(f *os.File, prefix string) *Logger {
	l := log.New(NewWriter(f), prefix, log.LstdFlags)
	return &Logger{l, f}
}

func NewWriter(f *os.File) *Writer {
	return &Writer{f}
}

type Writer struct {
	f *os.File
}

func (w *Writer) Write(p []byte) (int, error) {
	print("hi")
	n, err := os.Stdout.Write(p)
	if err != nil {
		return n, err
	}
	if n != len(p) {
		return n, io.ErrShortWrite
	}
	n, err = w.f.Write(p)
	if err != nil {
		return n, err
	}
	if n != len(p) {
		return n, io.ErrShortWrite
	}
	return len(p), nil
}

func isValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

func checkPasswordFormat(pw string) bool {
	exp := regexp.MustCompile("(?i)^[a-f0-9]{64}(:.+)?$")
	return exp.MatchString(pw)
}

func (u User) Claim() (claim *TokenClaim) {
	claim = new(TokenClaim)
	claim.Class = u.Class
	claim.ID = u.ID
	claim.Email = u.Email
	return
}

func fatal(err error) {
	panic(err.Error())
}

func UnauthorizedRequest(c echo.Context) error {
	response := new(Response)
	response.Success = false
	response.Error = "Unauthorized UserRequest"
	return c.JSON(http.StatusUnauthorized, response)
}
