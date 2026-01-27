# Troubleshooting

## Session Timeout

By default, user sessions expire after **2 hours**. If you're uploading large files over slower connections, you may need to increase this timeout to prevent sessions from expiring mid-upload. You can configure the session timeout using the `tokenExpirationTime` setting.

You can either set this option during runtime by using the flag `--tokenExpirationTime`, the environment variable `FB_TOKEN_EXPIRATION_TIME`, or in your configuration file. If you want to persist this to the configuration, please use [`filebrowser config set`](cli/filebrowser-config-set.md).

Valid duration formats include `"2h"`, `"30m"`, `"24h"`, or combinations like `"2h30m"`.
