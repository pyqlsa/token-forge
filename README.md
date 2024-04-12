# token-forge

A tool built out of curiousity, for research purposes only!

Inspired by GitHub's [blog](https://github.blog/2021-04-05-behind-githubs-new-authentication-token-formats/) about tokens and secret scanning.

## Why

> The internet is a scary place.

-- *Abraham Lincoln*

Why not?
- GitHub tokens have a schema; space of possible tokens is not `36^62`, rather `30^62`.
- GitHub has a lot of users and integrations; this means it is generating a lot of tokens; does a birthday attack become possible?
- GitHub's API `401`s upon failed authentication; GitHub's API rate limits also rise for authenticated clients.
- GitHub has an API to check your current rate limit; this API is not [rate-limited](https://docs.github.com/en/rest/overview/resources-in-the-rest-api?apiVersion=2022-11-28#checking-your-rate-limit-status-with-the-rest-api).
- Token expiration:
  - ghp: user-configurable expiration; forever tokens still supported.
  - gho: no apparent expiration (unless not used within the last year).
  - ghu: 8 hour expiration (or indefinite, depending on configuration).
  - ghs: 1 hour expiration.

Caveats:
- Mafs; big numbers are still big.
- [Secondary rate limits](https://docs.github.com/en/rest/overview/resources-in-the-rest-api?apiVersion=2022-11-28#secondary-rate-limits) exist to temporarily ban-hammer abusive entities.
  - I don't know what they are for `github.com`; values are undocumented and un-queryable.
  - They are disabled by default for [GitHub Enterprise Server](https://docs.github.com/en/enterprise-server@3.10/admin/configuration/configuring-user-applications-for-your-enterprise/configuring-rate-limits#enabling-secondary-rate-limits).
  - When enabled on GitHub Enterprise Server, the default values can be considered to be extremely high.

### Is this a bug?

Not according to GitHub. This was submitted as a bug bounty, and the response was:

> ...we are aware of the behavior you are describing and consider it to be an abuse issue and not a security vulnerability. We take abuse and spam seriously and have a dedicated team that tracks down spammy users. As a result, this is not eligible for reward under the Bug Bounty program.

## Warning

Tests for token validity are conducted against `github.com` by default. If running these tests against `github.com`, GitHub will very likely consider this interaction abuse. Proceed at your own risk!

This tool supports interacting with self-hosted GitHub Enterprise (by supplying the desired host as an argument). It is highly recommended to test against a GitHub Enterprise instance that you own or otherwise have authorization to test against.


## Usage

<!-- readme-help -->
```
Usage: token-forge <command> [flags]

A tool to 'work' with GitHub tokens.

Commands:
  version           Print version and exit.

  generate (gen)    Generate GitHub-like tokens.

  disect (dis)      Disect GitHub-like tokens.

  login             Test login with one or more tokens.

  local             Perform a local collision test.

  ip-check (ip)     Check resolved public ip address.

Flags:
  -h, --help    Show context-sensitive help.

Run "token-forge <command> --help" for more information on a command.
```
```
Usage: token-forge version [flags]

Print version and exit.

Flags:
  -h, --help    Show context-sensitive help.
```
```
Usage: token-forge generate (gen) [flags]

Generate GitHub-like tokens.

Flags:
  -h, --help     Show context-sensitive help.

      --debug    Enable debug mode

Token Params
  -b, --batch-size=1000    When testing for collisions, the number of tokens to
                           test concurrently.
  -n, --num-tokens=1       Number of tokens to test.
  -p, --prefix=STRING      Token prefix to use; if not specified, each generated
                           token will have a randomly selected prefix; only has
                           an effect when generating tokens.
```
```
Usage: token-forge disect (dis) --token=STRING --file=STRING --generated --no-auth [flags]

Disect GitHub-like tokens.

Flags:
  -h, --help     Show context-sensitive help.

      --debug    Enable debug mode

Source
  -t, --token=STRING    Token to use.
  -f, --file=STRING     Path to file with tokens.
  -g, --generated       Use one or more generated tokens.
  -x, --no-auth         Simply interact w/ the rate limit api with an
                        unauthenticated client.

Token Params
  -b, --batch-size=1000    When testing for collisions, the number of tokens to
                           test concurrently.
  -n, --num-tokens=1       Number of tokens to test.
  -p, --prefix=STRING      Token prefix to use; if not specified, each generated
                           token will have a randomly selected prefix; only has
                           an effect when generating tokens.
```
```
Usage: token-forge login --token=STRING --file=STRING --generated --no-auth [flags]

Test login with one or more tokens.

Flags:
  -h, --help           Show context-sensitive help.

      --debug          Enable debug mode
  -c, --force-check    Force a check of the logged in user so the rate limit is
                       decremented.
      --host=STRING    The GitHub Enterprise hostname to interact with; if not
                       specified, github.com is assumed.

Source
  -t, --token=STRING    Token to use.
  -f, --file=STRING     Path to file with tokens.
  -g, --generated       Use one or more generated tokens.
  -x, --no-auth         Simply interact w/ the rate limit api with an
                        unauthenticated client.

Token Params
  -b, --batch-size=1000    When testing for collisions, the number of tokens to
                           test concurrently.
  -n, --num-tokens=1       Number of tokens to test.
  -p, --prefix=STRING      Token prefix to use; if not specified, each generated
                           token will have a randomly selected prefix; only has
                           an effect when generating tokens.

Proxy Config
  --proxy=STRING    Proxy to use for outbound connections.
```
```
Usage: token-forge local [flags]

Perform a local collision test.

Flags:
  -h, --help           Show context-sensitive help.

      --debug          Enable debug mode
  -t, --num-tests=1    Number of tokens to load into the test token database.

Token Params
  -b, --batch-size=1000    When testing for collisions, the number of tokens to
                           test concurrently.
  -n, --num-tokens=1       Number of tokens to test.
  -p, --prefix=STRING      Token prefix to use; if not specified, each generated
                           token will have a randomly selected prefix; only has
                           an effect when generating tokens.
```
```
Usage: token-forge ip-check (ip) [flags]

Check resolved public ip address.

Flags:
  -h, --help     Show context-sensitive help.

      --debug    Enable debug mode

Proxy Config
  --proxy=STRING    Proxy to use for outbound connections.
```
<!-- readme-help end -->

### proxy

Breadcrumbs for a minimal local tor proxy are provided in the `./proxy` folder.

Build the proxy container:

```bash
./proxy/build.sh
```

Run the proxy container:

```bash
./proxy/run.sh
```

Extra private configurations can be placed in the `./proxy/priv.d` folder (e.g. bridges). All files in this folder will be copied into the container in a place where tor can detect them. If new files are added to this folder, or existing files are modified, the container must be rebuilt with `./proxy/build.sh` to be imported into the container image.

## Dev

This code contains a lot of hackery, but it seems to work.

### lint

Because you have to have at least *some* standards.

```bash
./scripts/lint.sh
```

### build

Build output for each `<platform>-<arch>` is written to the `./build` folder.

```bash
./scripts/build.sh
```

### test

```bash
./scrips/test.sh
# and
./scripts/test.sh race
```

### readme usage generation

```bash
./scripts/readme.sh
```
