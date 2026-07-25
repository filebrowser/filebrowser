# Security Policy

## Supported Versions

| Version | Supported |
| ------- | --------- |
| 2.x     | ✅         |
| < 2.0   | ❌         |

## Before Reporting

This project is in maintenance-only mode. To avoid duplicates, first check the [existing advisories](https://github.com/filebrowser/filebrowser/security/advisories) and open issues, and confirm:

- **It concerns this project, not a fork.** Reports about code, features, or endpoints that don't exist here belong to the relevant fork.
- **It isn't an already-known class** that remains unaddressed:
  - Command execution, runner, and hooks (opt-in, disabled by default) — [#5199](https://github.com/filebrowser/filebrowser/issues/5199)
  - Session and JWT handling — [#5216](https://github.com/filebrowser/filebrowser/issues/5216)

Reports covering these are likely to be closed as duplicates.

## Reporting a Vulnerability

- **Critical:** report privately via the [Security](https://github.com/filebrowser/filebrowser/security) page.
- **Non-critical:** open a public issue so the community can help; we'll label it as a security issue.

Please include, where possible:

- The commit the issue was found at
- A plaintext proof of concept (no binaries)
- Steps to reproduce
- Recommended remediation, if any

We're a volunteer effort, so responses can take a while, and we may reach out for clarification.
