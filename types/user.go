package types

import (
	"strings"

	"github.com/spf13/afero"
	"golang.org/x/crypto/bcrypt"
)

// ViewMode describes a view mode.
type ViewMode string

const (
	ListViewMode   ViewMode = "list"
	MosaicViewMode ViewMode = "mosaic"
)

// Permissions describe a user's permissions.
type Permissions struct {
	Admin    bool     `json:"admin"`
	Execute  bool     `json:"execute"`
	Create   bool     `json:"create"`
	Rename   bool     `json:"rename"`
	Modify   bool     `json:"edit"`
	Delete   bool     `json:"delete"`
	Share    bool     `json:"share"`
	Download bool     `json:"download"`
	Commands []string `json:"commands"`
}

// User describes a user.
type User struct {
	ID       uint        `storm:"id,increment" json:"id"`
	Username string      `storm:"unique" json:"username"`
	Password string      `json:"password"`
	Scope    string      `json:"scope"`
	Locale   string      `json:"locale"`
	ViewMode ViewMode    `json:"viewMode"`
	Perm     Permissions `json:"perm"`
	Fs       afero.Fs    `json:"-"`
	Rules    []Rule      `json:"rules"`
}

// BuildFs builds the FileSystem property of the user,
// which is the only one that can't be directly stored.
func (u *User) BuildFs() {
	if u.Fs == nil {
		u.Fs = afero.NewBasePathFs(afero.NewOsFs(), u.Scope)
	}
}

// IsAllowed checks if an user is allowed to go to a certain path.
func (u User) IsAllowed(url string) bool {
	var rule *Rule
	i := len(u.Rules) - 1

	for i >= 0 {
		rule = &u.Rules[i]

		if rule.Regex {
			if rule.Regexp.MatchString(url) {
				return rule.Allow
			}
		} else if strings.HasPrefix(url, rule.Path) {
			return rule.Allow
		}

		i--
	}

	return true
}

// HashPwd hashes a password.
func HashPwd(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPwd checks if a password is correct.
func CheckPwd(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
