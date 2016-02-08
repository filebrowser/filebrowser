# Hugo plugin for Caddy

[![Build](https://img.shields.io/travis/hacdias/caddy-hugo.svg?style=flat-square)](https://travis-ci.org/hacdias/caddy-hugo)
[![Documentation](https://img.shields.io/badge/caddy-doc-F06292.svg?style=flat-square)](https://caddyserver.com/docs/hugo)
[![Documentation](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](http://godoc.org/github.com/hacdias/caddy-hugo)

**Caddy-hugo fills the gap between Hugo and the browser.** [Hugo](http://gohugo.io/) is a easy, blazing fast and awesome static website generator. This plugin fills the gap between Hugo and the end-user, providing you an web interface to manage the whole website.

The following information is directed to those who want to build the plugin from source and make changes to it. If you just want to try it out, read the [documentation](https://caddyserver.com/docs/hugo) at Caddy website.

## Build from source

### Requirements

+ [Go 1.4 or higher][1]
+ [caddydev][3]
+ [go-bindata][5]
+ [Node.js w/ npm][6] (optional)


If you want to go deeper and make changes in front-end assets like JavaScript or CSS, you'll need to install the optional tools listed above. 

### Get it and build

1. Open the terminal.
2. Run ```go get github.com/hacdias/caddy-hugo```.
3. Navigate to the clone path.
4. Run ```go generate```.
  + If you want to make changes in the front-end, run ```go-bindata -debug -pkg assets -o assets/assets.go templates/ assets/css/ assets/js/ assets/fonts/``` too; execute ```npm install``` in the root of ```caddy-hugo``` clone. Then, run ```grunt watch```.
5. Open the folder with your static website and create a Caddyfile. Read the [docs](http://caddyserver.com/docs/hugo) for more information about the directives of this plugin.
6. Open the console in that folder and execute ```caddydev --source $PATH$ hugo```, replacing ```$PATH``` with the absolute path to your caddy-hugo's clone.
7. Open the browser and go to ```http://whateveryoururlis/admin``` to check it out.

[1]: https://golang.org/dl/
[2]: https://www.ruby-lang.org/en/
[3]: https://github.com/caddyserver/caddydev
[4]: http://sass-lang.com/install
[5]: https://github.com/jteeuwen/go-bindata
[6]: https://nodejs.org
[7]: http://gruntjs.com/
