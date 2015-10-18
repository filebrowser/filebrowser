# Hugo plugin for Caddy

[![Build](https://img.shields.io/travis/hacdias/caddy-hugo.svg?style=flat-square)](https://travis-ci.org/hacdias/caddy-hugo)
[![Documentation](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](http://godoc.org/github.com/hacdias/caddy-hugo)

Powerful [Hugo](http://gohugo.io/) - easy and amazing static website generator - plugin for Caddy with an admin interface so you can change your website when you're not on your computer. You can also use it like any other Content Management Service.

## Build it from source

### Requirements

| Back-end              | Front-end            |
| --------------------- | -------------------- |
| [Go 1.4 or higher][1] | [Ruby][2]            |
| [caddydev][3]         | [SASS][4]            |
| [go-bindata][5]       | [Node.js w/ npm][6]  |
|                       | [Grunt][7]           |

If you want to go deeper and make changes in front-end assets like JavaScript or CSS, you'll need some more tools (front-end tools in the table bellow). If you don't, install only the back-end tools.

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
