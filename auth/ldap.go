package auth

import (
    "net/http"
    "encoding/json"
    "crypto/tls"
    "fmt"
    "strings"

    "github.com/filebrowser/filebrowser/v2/settings"
    "github.com/filebrowser/filebrowser/v2/users"
    "github.com/filebrowser/filebrowser/v2/errors"
    "github.com/go-ldap/ldap/v3"
)

const MethodLDAPAuth settings.AuthMethod = "ldap"

type LDAPAuth struct {
    Server     string `json:"server"`
    StartTLS   bool   `json:"starttls"`
    SkipVerify bool   `json:"skipverify"`
    BaseDN     string `json:"basedn"`
    UserOU     string `json:"userou"`
    GroupOU    string `json:"groupou"`
    UserCN     string `json:"usercn"`
    AdminCN    string `json:"admincn"`
    UserHome   string `json:"userhome"`
}

func (a LDAPAuth) Auth(r *http.Request, sto *users.Storage, set *settings.Storage, root string) (user *users.User, err error) {
    var cred jsonCred

    if r.Body == nil {
        return nil, errors.ErrEmptyRequest
    }

    err = json.NewDecoder(r.Body).Decode(&cred)
    if err != nil {
        return nil, err
    }

    p := strings.Split(a.Server, ":")
    var l *ldap.Conn
    if p[0] == "ldaps" {
        l, err = ldap.DialURL(a.Server, ldap.DialWithTLSConfig(&tls.Config{InsecureSkipVerify: a.SkipVerify}))
    } else {
        l, err = ldap.DialURL(a.Server)
        if err != nil {
            return nil, err
        }
        if a.StartTLS {
            err = l.StartTLS(&tls.Config{InsecureSkipVerify: a.SkipVerify})
        }
    }
    if err != nil {
        return nil, err
    }
    bind := fmt.Sprintf("uid=%s,ou=%s,%s", cred.Username, a.UserOU, a.BaseDN)
    err = l.Bind(bind, cred.Password)
    if err != nil {
        return nil, err
    }

    var s *settings.Settings
    s, err = set.Get()
    if err != nil {
        return nil, err
    }

    var isadmin bool
    var hashome string = s.Defaults.Scope
    if a.AdminCN != "" || a.UserCN != "" || a.UserHome != "" {
        searchRequest := ldap.NewSearchRequest(
            bind,
            ldap.ScopeWholeSubtree,
            ldap.NeverDerefAliases,
            0,
            0,
            false,
            "(&(objectClass=*))",
            []string{"memberOf", a.UserHome},
            nil,
        )
        searchResult, err := l.Search(searchRequest)
        if err != nil {
            return nil, err
        }
        l.Close()

        var isuser bool
        admingrp := fmt.Sprintf("cn=%s,ou=%s,%s", a.AdminCN, a.GroupOU, a.BaseDN)
        usergrp := fmt.Sprintf("cn=%s,ou=%s,%s", a.UserCN, a.GroupOU, a.BaseDN)
        for _, group := range searchResult.Entries[0].GetAttributeValues("memberOf") {
            switch group {
                case admingrp:
                    isadmin = true
                case usergrp:
                    isuser = true 
            }
        }
        if a.UserHome != "" {
            hashome = searchResult.Entries[0].GetAttributeValue(a.UserHome)
        }

        // Deny entry to non-users if user group is enabled, admins always have access
        if !isadmin && a.UserCN != "" && !isuser {
            return nil, errors.ErrPermissionDenied
        }
    }

    user, err = sto.Get(root, cred.Username)
    if err != nil {
        if err == errors.ErrNotExist {
            user = &users.User{
                Username: cred.Username,
                Password: "much5af3v3rys3cur3", // No point hashing the password since we don't use it
                LockPassword: true,             // Prevent user password change which would only lead to confusion
            }
            s.Defaults.Apply(user)
            user.Perm.Admin = isadmin
            if user.Scope != hashome {
                user.Scope = hashome
            } else {
                home, err := s.MakeUserDir(cred.Username, user.Scope, root)
                if err != nil {
                    return nil, err
                }
                user.Scope = home
            }
            err = sto.Save(user)
            if err != nil {
                return nil, err
            }
        } else {
            return nil, err
        }
    } else {
        // Keep profile in sync with LDAP
        var update bool
        if user.Perm.Admin != isadmin {
            user.Perm.Admin = isadmin
            update = true
        }
        if user.Scope != hashome {
            user.Scope = hashome
            update = true
        }
        if update {
            sto.Update(user)
        }
    }

    return
}

func (a LDAPAuth) LoginPage() bool {
    return true
}