# Hugo add-on for Caddy

This is an add-on for Caddy which wants to deliver a good UI to edit the content of the website.

## Try it

You have to instal ```go-bindata``` before. Then execute the following command:

```
go-bindata -debug -pkg assets -o assets/assets.go templates/ assets/dist/css/ assets/dist/js/
```

Now you're ready to test it using Caddydev.
