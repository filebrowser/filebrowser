package filemanager

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

var defaultCredentials = "{\"username\":\"admin\",\"password\":\"admin\"}"

var authHandlerTests = []struct {
	Data     string
	Expected int
}{
	{defaultCredentials, http.StatusOK},
	{"{\"username\":\"admin\",\"password\":\"wrong\"}", http.StatusForbidden},
	{"{\"username\":\"wrong\",\"password\":\"admin\"}", http.StatusForbidden},
}

func TestAuthHandler(t *testing.T) {
	fm := newTest(t)
	defer fm.Clean()

	for _, test := range authHandlerTests {
		req, err := http.NewRequest("POST", "/api/auth/get", strings.NewReader(test.Data))
		if err != nil {
			t.Fatal(err)
		}

		w := httptest.NewRecorder()
		fm.ServeHTTP(w, req)

		if w.Code != test.Expected {
			t.Errorf("Wrong status code: got %v want %v", w.Code, test.Expected)
		}
	}
}

func TestRenewHandler(t *testing.T) {
	fm := newTest(t)
	defer fm.Clean()

	// First, we have to make an auth request to get the user authenticated,
	r, err := http.NewRequest("POST", "/api/auth/get", strings.NewReader(defaultCredentials))
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	fm.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Couldn't authenticate: got %v", w.Code)
	}

	token := w.Body.String()

	// Test renew authorization via Authorization Header.
	r, err = http.NewRequest("GET", "/api/auth/renew", nil)
	if err != nil {
		t.Fatal(err)
	}

	r.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	fm.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Can't renew auth via header: got %v", w.Code)
	}

	// Test renew authorization via cookie field.
	r, err = http.NewRequest("GET", "/api/auth/renew", nil)
	if err != nil {
		t.Fatal(err)
	}

	r.AddCookie(&http.Cookie{
		Value:   token,
		Name:    "auth",
		Expires: time.Now().Add(1 * time.Hour),
	})

	w = httptest.NewRecorder()
	fm.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Can't renew auth via cookie: got %v", w.Code)
	}
}
