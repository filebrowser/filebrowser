package filemanager

import (
	"net/http"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/dgrijalva/jwt-go/request"
)

/* Set up a global string for our secret */
var key = []byte("secret")

type claims struct {
	*User
	jwt.StandardClaims
}

func getTokenHandler(c *requestContext, w http.ResponseWriter, r *http.Request) (int, error) {
	// TODO: get user and password info from the request
	// check if the password is correct for that user using a DB or JSOn
	// or something.

	claims := claims{
		c.fm.User,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Minute * 5).Unix(),
			Issuer:    "test",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	string, err := token.SignedString(key)

	if err != nil {
		return http.StatusInternalServerError, err
	}

	w.Write([]byte(string))
	return 0, nil
}

func validAuth(c *requestContext, r *http.Request) (bool, *User) {
	token, err := request.ParseFromRequestWithClaims(r, request.AuthorizationHeaderExtractor, &claims{},
		func(token *jwt.Token) (interface{}, error) {
			return key, nil
		})

	if err == nil && token.Valid {
		return true, c.fm.User
	}

	return false, nil
}
