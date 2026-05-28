package share

import (
	"reflect"
	"testing"
)

type fakeBackend struct {
	links []*Link
}

func (f fakeBackend) All() ([]*Link, error) {
	return f.links, nil
}

func (f fakeBackend) FindByUserID(id uint) ([]*Link, error) {
	var links []*Link
	for _, link := range f.links {
		if link.UserID == id {
			links = append(links, link)
		}
	}
	return links, nil
}

func (f fakeBackend) GetByHash(hash string) (*Link, error) {
	for _, link := range f.links {
		if link.Hash == hash {
			return link, nil
		}
	}
	return nil, nil
}

func (f fakeBackend) GetPermanent(path string, id uint) (*Link, error) {
	for _, link := range f.links {
		if link.Path == path && link.UserID == id && link.Expire == 0 {
			return link, nil
		}
	}
	return nil, nil
}

func (f fakeBackend) Gets(path string, id uint) ([]*Link, error) {
	var links []*Link
	for _, link := range f.links {
		if link.Path == path && link.UserID == id {
			links = append(links, link)
		}
	}
	return links, nil
}

func (f fakeBackend) Save(_ *Link) error {
	return nil
}

func (f fakeBackend) Delete(_ string) error {
	return nil
}

func (f fakeBackend) DeleteWithPathPrefix(_ string) error {
	return nil
}

func TestGetsByPathReturnsLinksFromAllUsers(t *testing.T) {
	t.Parallel()

	expected := []*Link{
		{Hash: "a", Path: "/file.txt", UserID: 1},
		{Hash: "b", Path: "/file.txt", UserID: 2},
	}
	store := NewStorage(fakeBackend{
		links: append(expected, &Link{Hash: "c", Path: "/other.txt", UserID: 3}),
	})

	links, err := store.GetsByPath("/file.txt")
	if err != nil {
		t.Fatalf("GetsByPath returned error: %v", err)
	}

	if !reflect.DeepEqual(links, expected) {
		t.Fatalf("GetsByPath returned %#v, want %#v", links, expected)
	}
}
