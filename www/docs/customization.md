# Customization

You can customize the styles, branding and icons of your File Browser instance in order to give it a personal touch.

## Custom Branding

You can customize File Browser to use your own branding. This includes the following:

- **Name**: the name of the instance that shows up on the tab title, login pages, and some other places.
- **Disable External Links**: disables all external links, except to the documentation.
- **Disable Used Percentage**: disables the disk usage information on the sidebar.
- **Branding Folder**: directory which can contain two items:
  - `custom.css`, containing a global stylesheet to apply to all users.
  - `img`, a directory which can replace all the [default logotypes](https://github.com/filebrowser/filebrowser/tree/master/frontend/public/img) from the application.

This can be configured by the administrator user, under **Settings → Global Settings**. You can also update the configuration directly using the [CLI](cli/filebrowser-config-set.md):

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
