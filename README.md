# caddy-hugo

[![Build](https://img.shields.io/travis/hacdias/caddy-hugo.svg?style=flat-square)](https://travis-ci.org/hacdias/caddy-hugo)
[![Documentation](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](http://godoc.org/github.com/hacdias/caddy-hugo)

Powerful and easy static site generator with admin interface with [Hugo](http://gohugo.io/) (and you don't need to install it separately).

## Build it from source

If you want to try caddy-hugo plugin (and improve it maybe), you'll have to install some tools.

+ [Go 1.4 or higher](https://golang.org/dl/)
+ [caddydev](https://github.com/caddyserver/caddydev)
+ [go-bindata](https://github.com/jteeuwen/go-bindata)

If you want to go deeper and make changes in front-end assets like JavaScript or CSS, you'll need some more tools.

+ [Ruby](https://www.ruby-lang.org/en/)
+ [SASS](http://sass-lang.com/install)
+ [Node.js and npm](https://nodejs.org)
+ [Grunt](http://gruntjs.com/)

### Run it

If you have already installed everything above to meet the requirements for what you want to do, let's start. Firstly, open the terminal and navigate to your clone of ```caddy-hugo```. Then execute:

```
go-bindata [-debug] -pkg assets -o assets/assets.go templates/ assets/css/ assets/js/ assets/fonts/
```

That command will create an ```assets.go``` file which contains all static files from those folders mentioned in the command. You may run with ```-debug``` option if you want, but it is only needed if you're going to make changes in front-end assets.

Now, open the folder with your static website and create a Caddyfile. Read the [docs](http://caddyserver.com/docs/cms) for more information about the directives of this plugin.

After creating the file, navigate to that folder using the terminal and run the following command, replacing ```{caddy-hugo}``` with the location of your clone.

```
caddydev --source {caddy-hugo} hugo
```

Navigate to the url you set on Caddyfile to see your blog running on Caddy and Hugo. Go to ```/admin``` to try the Admin UI.

Everything is working now. Whenever you make a change in the back-end source code, you'll have to run the command above again.

**For those who want to make changes in front-end**, make sure you have every needed tool installed and run ```npm install``` in the root of ```caddy-hugo``` clone. Then, run ```grunt watch```.
