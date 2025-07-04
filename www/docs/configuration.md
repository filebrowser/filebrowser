# Configuration

Most of the configuration can be understood through our Command Line Interface documentation. Although there are some specific topics that we want to cover on this section.

## Custom Branding

You can customize File Browser to use your own branding. This includes the following:

- **Name**: the name of the instance that shows up on the tab title, login pages, and some other places.
- **Disable External Links**: disables all external links, except to the documentation.
- **Disable Used Percentage**: disables the disk usage information on the sidebar.
- **Branding Folder**: directory which can contain two items:
    - `custom.css`, containing a global stylesheet to apply to all users.
    - `img`, a directory which can replace all the [default logotypes](https://github.com/filebrowser/filebrowser/tree/master/frontend/public/img) from the application.

This can be configured by the administrator user, under **Settings → Global Settings**. You can also update the configuration directly using the CLI:

```sh
filebrowser config set --branding.name "My Name" \
  --branding.files "/abs/path/to/my/dir" \
  --branding.disableExternal
```

> [!NOTE] 
>
> If you are using Docker, you need to mount a volume with the `branding` directory in order for it to be accessible from within the container.

### Custom Icons

To replace the default logotype and favicons, you need to create an `img` directory under the branding directory. The structure of this directory must mimic the one from the [default logotypes](https://github.com/filebrowser/filebrowser/tree/master/frontend/public/img):

```
img/
  logo.svg
  icons/
    favicon.ico
    favicon.svg
    (...)
```

Note that there are different versions of the same favicon in multiple sizes. To replace all of them, you need to add versions for all of them. You can use the [Real Favicon Generator](https://realfavicongenerator.net/) to generate these for you from your base image. 

> [!NOTE]
>
> The icons are cached by the browser, so you may not see your changes immediately. You can address this by clearing your browser's cache.

## Authentication Method

Right now, there are three possible authentication methods. Each one of them has its own capabilities and specification. If you are interested in contributing with one more authentication method, please [check the guidelines](contributing.md).

### JSON Auth (default)

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

### Proxy Header

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

## Command Runner

> [!CAUTION]
>
> The **command execution** functionality has been disabled for all existent and new installations by default from version v2.33.8 and onwards, due to continuous and known security vulnerabilities. You should only use this feature if you are aware of all of the security risks involved. For more up to date information, consult issue [#5199](https://github.com/filebrowser/filebrowser/issues/5199).

The command runner is a feature that enables you to execute any shell command you want before or after a certain event. Right now, these are the events:

* Copy
* Rename
* Upload
* Delete
* Save

Also, during the execution of the commands set for those hooks, there will be some environment variables available to help you perform your commands:

* `FILE` with the full absolute path to the changed file.
* `SCOPE` with the path to user's scope.
* `TRIGGER` with the name of the event.
* `USERNAME` with the user's username.
* `DESTINATION` with the absolute path to the destination. Only used for **copy** and **rename.**

At this moment, you can edit the commands via the command line interface, using the following commands \(please check the flag `--help` to know more about them\):

```bash
filebrowser cmds add before_copy "echo $FILE"
filebrowser cmds rm before_copy 0
filebrowser cmds ls
```

Or you can use the web interface to manage them via **Settings** → **Global Settings**.

## Command Execution

> [!CAUTION]
>
> The **command execution** functionality has been disabled for all existent and new installations by default from version v2.33.8 and onwards, due to continuous and known security vulnerabilities. You should only use this feature if you are aware of all of the security risks involved. For more up to date information, consult issue [#5199](https://github.com/filebrowser/filebrowser/issues/5199).

Within File Browser you can toggle the shell (`< >` icon at the top right) and this will open a shell command window at the bottom of the screen. This functionality can be turned on using the environment variable `FB_DISABLE_EXEC=false` or the flag `--disable-exec=false`.

By default no commands are available as the command list is empty. To enable commands these need to either be done on a per-user basis (including for the Admin user).

You can do this by adding them in Settings > User Management > (edit user) > Commands or to *apply to all new users created from that point forward* they can be set in Settings > Global Settings

> [!NOTE]
> 
> If using a proxy manager then remember to enable websockets support for the File Browser proxy

> [!NOTE]
> 
> If using Docker and you want to add a new command that is not in the base image then you will need to build a custom Docker image using `filebrowser/filebrowser` as a base image.  For example to add 7z:
> 
> ```docker
> FROM filebrowser/filebrowser
> RUN sudo apt install p7zip-full
> ```
