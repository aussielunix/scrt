---
description: Learn to configure scrt, a command-line secret manager for developers, sysadmins and devops. Use configuration files and environment variables to configure scrt for usage in your projects.
---

# Configuration

Repeating the global options each time the `scrt` command is invoked can be verbose. Also, some options — like the store password — shouldn't be used on the command line on a shared computer, to avoid security issues.

To prevent this, scrt can be configured with a configuration file or using environment variables.

When you do want to provide the password on the command line, `--password` can also be passed without a value. In that case, `scrt` will prompt for the password interactively and hide the typed characters.

`scrt` configuration settings follow an order of precedence. Each item takes precedence over the item below it:

- command-line flags
- environment variables
- configuration file

::: tip
Configuration options can be considered to be chosen from "most explicit" (flags) to "least explicit" (configuration file).
:::

## Configuration file

The `scrt` configuration file is a [YAML](https://yaml.org/) file with the configuration options as keys.

Example:

```yaml
storage: local
password: p4ssw0rd
local:
  path: store.scrt
```

For Git storage, configure both the repository URL and the local clone path:

```yaml
storage: git
password: p4ssw0rd
git:
  url: git@github.com:githubuser/secrets.git
  path: store.scrt
  local-path: ~/.scrt/repos/secrets
```

If the `--config` option is given to the command line, `scrt` will try to load the configuration from a file at the given path. Otherwise, it first looks for `config.yml` in the current working directory. If that file does not exist, it then looks for `~/.scrt/config.yml`.

You can also bootstrap a new configuration file with `scrt init --config=...`. When `init` succeeds, `scrt` will create or overwrite the config file and populate it with the options that were explicitly provided as CLI flags.

This can be useful in configuring the location of a store for a project. By adding a `.scrt` file at the root of the project repository. `scrt` can then be used in CI and other DevOps tools.

::: danger
Don't add the password to a configuration file in a shared git repository!
:::

Storage type (`storage`) can be ignored in a configuration file. `scrt` will read the configuration under the key for the storage type (e.g. `local:`). _Defining configurations for multiple storage types in a single file will result in undefined behavior._

## Environment variables

Each global option has an environment variable counterpart. Environment variables use the same name as the configuration option, in uppercase letters, prefixed with `SCRT_`.

- `storage` ⇒ `SCRT_STORAGE`
- `password` ⇒ `SCRT_PASSWORD`
- `local-path` ⇒ `SCRT_LOCAL_PATH`
- `git-local-path` ⇒ `SCRT_GIT_LOCAL_PATH`

To configure a default store on your system, add the following to your `.bashrc` file (if using `bash`):

```bash
export SCRT_STORAGE=local
export SCRT_PASSWORD=p4ssw0rd
export SCRT_LOCAL_PATH=~/.scrt/store.scrt
```

For Git storage, set the local clone path as well:

```bash
export SCRT_STORAGE=git
export SCRT_PASSWORD=p4ssw0rd
export SCRT_GIT_URL=git@github.com:githubuser/secrets.git
export SCRT_GIT_PATH=store.scrt
export SCRT_GIT_LOCAL_PATH=~/.scrt/repos/secrets
```

::: tip
Refer to your shell interpreter's documentation to set environment variables if you don't use `bash` (`zsh`, `dash`, `tcsh`, etc.)
:::

#### Related pages

- [Reference > Configuration](../reference/configuration/README.md)
