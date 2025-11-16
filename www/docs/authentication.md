# Authentication

There are three possible authentication methods. Each one of them has its own capabilities and specification. If you are interested in contributing with one more authentication method, please [check the guidelines](contributing.md).

## JSON Auth (default)

We call it JSON Authentication but it is just the default authentication method and the one that is provided by default if you don't make any changes. It is set by default, but if you've made changes before you can revert to using JSON auth:

```sh
filebrowser config set --auth.method=json
```

This method can also be extended with **reCAPTCHA** verification during login:

```sh
filebrowser config set --auth.method=json \
  --recaptcha.key site-key \
  --recaptcha.secret private-key
```

By default, we use [Google's reCAPTCHA](https://developers.google.com/recaptcha/docs/display) service. If you live in China, or want to use other provider, you can change the host with the following command:

```sh
filebrowser config set --recaptcha.host https://recaptcha.net
```

Where `https://recaptcha.net` is any provider you want.

## Proxy Header

If you have a reverse proxy you want to use to login your users, you do it via our `proxy` authentication method. To configure this method, your proxy must send an HTTP header containing the username of the logged in user:

```sh
filebrowser config set --auth.method=proxy --auth.header=X-My-Header
```

Where `X-My-Header` is the HTTP header provided by your proxy with the username.

> [!WARNING]
> 
> File Browser will blindly trust the provided header. If the proxy can be bypassed, an attacker could simply attach the header and get admin access.

### No Authentication

We also provide a no authentication mechanism for users that want to use File Browser privately such in a home network. By setting this authentication method, the user with **id 1** will be used as the default users. Creating more users won't have any effect.

```sh
filebrowser config set --auth.method=noauth
```

## Session Timeout

By default, user sessions expire after **2 hours**. If you're uploading large files over slower connections, you may need to increase this timeout to prevent sessions from expiring mid-upload. You can configure the session timeout using the `token-expiration-time` setting.

### Configuration File

Add the setting to your configuration file (e.g., `/config/settings.json` in Docker):

```json
{
  "token-expiration-time": "6h"
}
```

> [!IMPORTANT]
>
> The key must use kebab-case format: `token-expiration-time`. Valid duration formats include `"2h"`, `"30m"`, `"24h"`, or combinations like `"2h30m"`.

### Environment Variable

Set the corresponding environment variable:

```sh
docker run -e FB_TOKEN_EXPIRATION_TIME=6h ...
```

### CLI Flag

Pass the flag when starting File Browser:

```sh
filebrowser --token-expiration-time 6h
```

### Updating an Existing Installation

File Browser saves configuration values to the database during the **first run**.
Updating `settings.json` or environment variables later **will not affect an existing installation**.
To change the timeout, use:

```sh
filebrowser config set --token-expiration-time 6h
```
