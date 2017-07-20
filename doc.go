/*
Package filemanager provides a web interface to access your files
wherever you are. To use this package as a middleware for your app,
you'll need to create a filemanager instance:

	m, err := filemanager.New(database, user)

Where 'user' contains the default options for new users. You can just
use 'filemanager.DefaultUser' or create yourself a default user:

	m, err := filemanager.New(database, filemanager.User{
		Admin: 		   false,
		AllowCommands: false,
		AllowEdit:     true,
		AllowNew:      true,
		Commands:      []string{
			"git",
		},
		Rules:         []*filemanager.Rule{},
		CSS:           "",
		FileSystem:    webdav.Dir("/path/to/files"),
	})

The credentials for the first user are always 'admin' for both the user and
the password, and they can be changed later through the settings. The first
user is always an Admin and has all of the permissions set to 'true'.

Then, you should set the Prefix URL and the Base URL, using the following
functions:

	m.SetBaseURL("/")
	m.SetPrefixURL("/")

The Prefix URL is a part of the path that is already stripped from the
r.URL.Path variable before the request arrives to File Manager's handler.
This is a function that will rarely be used. You can see one example on Caddy
filemanager plugin.

The Base URL is the URL path where you want File Manager to be available in. If
you want to be available at the root path, you should call:

	m.SetBaseURL("/")

But if you want to access it at '/admin', you would call:

	m.SetBaseURL("/admin")

Now, that you already have a File Manager instance created, you just need to
add it to your handlers using m.ServeHTTP which is compatible to http.Handler.
We also have a m.ServeWithErrorsHTTP that returns the status code and an error.

One simple implementation for this, at port 80, in the root of the domain, would be:

	http.ListenAndServe(":80", m)
*/
package filemanager
