# Configuration

Most of the configuration can be understood through our Command Line Interface documentation. Although there are some specific topics that we want to cover on this section.

## Custom Branding

You are able to customize your File Browser installation by changing its name to any other you want, by adding a global custom style sheet and by using your own logotype if you want. To address this, there are three configuration options that can be changed:

* **Name:** which is the instance name that will show up on login and signup pages. This won't replace the version message in the sidebar.
* **Disable external links:** this will disable any external links (except the ones to this documentation).
* **Folder:** is the path to a directory that can contain two items:
  * **custom.css**, containing the styles you want to apply to your installation.
  * **img** a directory whose files can replace the [default logotypes](../frontend/public/img) in the application.

These options can be either set via the CLI interface using the following command:

```sh
filebrowser config set --branding.name "My Name" \
  --branding.files "/abs/path/to/my/dir" \
  --branding.disableExternal
```
Or can be set under 'Branding directory path' in **Settings → Global Settings**. 

> [!NOTE] 
>
> If using Docker then remember to bind this directory, for example as `/home/username/containers/filebrowser/branding:/branding`

For custom icons to be recognized you need to create `img` and `img/icons` directories and place the svg in the `branding/img` directory:

```
- filebrowser
  - branding
    - img
      - icons
      - logo.svg
  - filebrowser.db
```

To replace the favicon you need to place this in the `img/icons` directory but also note that some of the other PNG icon types will be required too (see the default logotypes link above) as the browser will normally use the highest resolution option available (at a minimum the 16x16 and 32x32 options). You can use the [Real Favicon Generator](https://realfavicongenerator.net/) to generate these for you from your base image.  

The icons are cached, to make the new ones appear more quickly open developer tools in your browser, then click on the Application tab, then Storage and then 'Clear Site Data'.

## Authentication Method

Right now, there are three possible authentication methods. Each one of them has its own capabilities and specification. If you are interested in contributing with one more authentication method, please [check the guidelines](./contributing.md).

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


> [!CAUTION]
> 
> Note that you **always** need to set the `--auth.method` flag when changing authentication configurations and that it will completely overwrite your current settings. [This is a known issue.](https://github.com/filebrowser/filebrowser/issues/715)

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


## Shell commands

Within Filebrowser you can toggle the shell (`< >` icon at the top right) and this will open a shell command window at the bottom of the screen.

**By default no commands are available as the command list is empty**

To enable commands these need to either be done on a per-user basis (including for the Admin user).

You can do this by adding them in Settings > User Management > (edit user) > Commands or to *apply to all new users created from that point forward* they can be set in Settings > Global Settings

> [!NOTE]
> 
> If using a proxy manager then remember to enable websockets support for the Filebrowser proxy

> [!NOTE]
> 
> If using Docker and you want to add a new command that is not in the base image then you will need to build a custom Docker image using `filebrowser/filebrowser` as a base image.  For example to add 7z:
> 
> ```docker
> FROM filebrowser/filebrowser
> RUN sudo apt install p7zip-full
> ```
