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

## Hook Authentication

The Hook Authentication method in FileBrowser allows developers to delegate user authentication to an external script or program. Instead of validating credentials internally, FileBrowser sends the username and password to a custom command defined by the administrator. This command receives the credentials through environment variables and returns key‑value pairs indicating whether the user should be authenticated, blocked, or passed through.

The hook’s output controls user permissions, scope, locale, and other attributes, making it a powerful and extensible authentication mechanism.

For example, the following code delegates filebrowser authentication to a PowerShell script on Windows. You can configure any command (for example, a script in Python, Node.js, etc.).

```sh
filebrowser config set --auth.method=hook --auth.command="powershell.exe -File C:\route\to\your\script\auth.ps1"
```

This is the code for the auth.ps1 script

```sh
param()

# Get FileBrowser credentials from environment variables
$username = $env:USERNAME
$password = $env:PASSWORD

# Users dictionary (for testing purposes only)
$users = @{
    "admin" = "kideW48v7-SdE*"
    "test"  = "2sDd3-etrytñK"
}

# Check if the user exists in the dictionary and verify the password
if ($users.ContainsKey($username) -and $users[$username] -eq $password) {

    # Successful authentication
    Write-Output "hook.action=auth"        # Hook action (in this case, "auth", is required for successful authentication)

    Write-Output "user.perm.admin=true"    # Set admin role (all permissions)
    #You can also define specific permissions like this:
    Write-Output "user.perm.execute=true"
    Write-Output "user.perm.create=true"
    Write-Output "user.perm.rename=true"
    Write-Output "user.perm.modify=true"
    Write-Output "user.perm.delete=true"
    Write-Output "user.perm.share=true"
    Write-Output "user.perm.download=true"

    Write-Output "user.locale=es"          # Set language
    Write-Output "user.viewMode=list"      # Set view mode
    Write-Output "user.scope=/"            # Set FileBrowser scope
    Write-Output "user.singleClick=true"   # Set single click user configuration
    Write-Output "user.hideDotfiles=false" # Set hide dot files user configuration

    #Set other configuration
} else {
    # Block authentication
    Write-Output "hook.action=block"
}
```

### Hook Output Format

A hook authentication script must output a series of key–value pairs, one per line, using the format:

```
key=value
```

FileBrowser reads these lines and applies the corresponding authentication action and user configuration.

#### Required Fields

The hook must output one of the following actions:

| Key    | Description |
|--------|------------ |
| hook.action=auth | Authenticates the user. FileBrowser will create or update the user if needed. |
| hook.action=block | Rejects authentication. The login attempt fails. |
| hook.action=pass | Delegates authentication to FileBrowser’s internal password validation. |

For most custom authentication flows, auth or block are used.

Example of a successful authentication:

```sh
hook.action=auth
```

#### Optional User Fields

When `hook.action=auth` is returned, the hook may also define additional user attributes. These fields override FileBrowser defaults and allow full customization of the authenticated user.

1. Permissions
```
user.perm.admin=true
user.perm.execute=true
user.perm.create=true
user.perm.rename=true
user.perm.modify=true
user.perm.delete=true
user.perm.share=true
user.perm.download=true
```
> Setting user.perm.admin=true automatically enables all permissions.

2. User Interface and Behavior
```
user.locale=es
user.viewMode=list
user.singleClick=true
user.hideDotfiles=false
```

3. User Scope
```
user.scope=/
```

## No Authentication

We also provide a no authentication mechanism for users that want to use File Browser privately such in a home network. By setting this authentication method, the user with **id 1** will be used as the default users. Creating more users won't have any effect.

```sh
filebrowser config set --auth.method=noauth
```
