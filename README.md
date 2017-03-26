# hugo - a caddy plugin

[![Build](https://img.shields.io/travis/hacdias/caddy-hugo.svg?style=flat-square)](https://travis-ci.org/hacdias/caddy-hugo)
[![community](https://img.shields.io/badge/community-forum-ff69b4.svg?style=flat-square)](https://forum.caddyserver.com)
[![Documentation](https://img.shields.io/badge/caddy-doc-F06292.svg?style=flat-square)](https://caddyserver.com/docs/hugo)
[![Documentation](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](http://godoc.org/github.com/hacdias/caddy-hugo)


hugo fills the gap between Hugo and the browser. [Hugo](http://gohugo.io/) is an easy and fast static website generator. This plugin fills the gap between Hugo and the end-user, providing you an web interface to manage the whole website.

Using this plugin, you won't need to have your own computer to edit posts, neither regenerate your static website, because you can do all of that just through your browser.

**Requirements:** you need to have the hugo executable in your PATH. You can download it from its [official page](http://gohugo.io).

### Syntax

```
hugo [directory] [admin] {
    flag name [value]
    # other file manager compatible options
}
```

All of the options above are optional.

* **directory** is the folder where the commands are going to be executed. By default, it is the current working directory. Default: `./`.
* **admin** is the path where you will find your administration interface. Default: `/admin`.
* **name** refers to the Hugo available flags. Please use their long form without `--` in the beginning. If no **value** is set, it will be evaluated as `true`.

In spite of these options, you can also use the [filemanager](/docs/filemanager) so you can have more control about what can be acceded, the permissions of each user, and so on.

This directive should be used with [root](/docs/root), [basicauth](/docs/basicauth) and [errors](/docs/errors) middleware to have the best experience. See the examples to know more.

### Examples

If you don't already have an Hugo website, don't worry. This plugin will auto-generate it for you. But that's not everything. It is recommended that you take a look at Hugo [documentation](http://gohugo.io/themes/overview/) to learn more about themes, content types, and so on.

A simple Caddyfile to use with Hugo static website generator:

```
root      public           # the folder where Hugo generates the website
basicauth /admin user pass # protect the admin area using HTTP basic auth
hugo                       # enable the admin panel
```
