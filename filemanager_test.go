package filemanager

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/net/webdav"
)

type test struct {
	*FileManager
	Temp string
}

func (t test) Clean() {
	t.db.Close()
	os.RemoveAll(t.Temp)
}

func newTest(t *testing.T) *test {
	temp, err := ioutil.TempDir("", t.Name())
	if err != nil {
		t.Fatalf("Error creating temporary directory: %v", err)
	}

	scope := filepath.Join(temp, "scope")
	database := filepath.Join(temp, "database.db")

	err = copyDir("./testdata", scope)
	if err != nil {
		t.Fatalf("Error copying the test data: %v", err)
	}

	user := DefaultUser
	user.FileSystem = webdav.Dir(scope)

	fm, err := New(database, user)

	if err != nil {
		t.Fatalf("Error creating a file manager instance: %v", err)
	}

	return &test{
		FileManager: fm,
		Temp:        temp,
	}
}
