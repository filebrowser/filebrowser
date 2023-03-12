package audit

import (
	"bytes"
	"github.com/filebrowser/filebrowser/v2/users"
	"log"
	"testing"
)

func TestLogFileActivity(t *testing.T) {
	resultBuffer := setupTestLogger()

	info := ResourceActivity{
		Event:        "Deletion",
		ResourcePath: "/srv/test.txt",
		User: &users.User{
			Username: "test",
			ID:       42,
		},
	}

	LogResourceActivity(info)

	result := resultBuffer.String()
	expectedResult := "Deletion of resource with path /srv/test.txt by user test (42)\n"
	if result != expectedResult {
		t.Errorf("Log entry should be \"%v\" but is \"%v\"", expectedResult, result)
	}
}

func TestLogFileActivityDirectory(t *testing.T) {
	resultBuffer := setupTestLogger()

	info := ResourceActivity{
		Event:        "Creation",
		ResourcePath: "/srv/test.txt",
		User: &users.User{
			Username: "test",
			ID:       42,
		},
	}

	LogResourceActivity(info)

	result := resultBuffer.String()
	expectedResult := "Creation of resource with path /srv/test.txt by user test (42)\n"
	if result != expectedResult {
		t.Errorf("Log entry should be \"%v\" but is \"%v\"", expectedResult, result)
	}
}

func setupTestLogger() *bytes.Buffer {
	var testBuffer bytes.Buffer
	testLogger := log.New(&testBuffer, "", 0)

	auditLogger = testLogger

	return &testBuffer
}
