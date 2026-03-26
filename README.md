# envcheck

A CLI tool that keeps your `.env` files honest.

![envcheck demo](demo.gif)

## What it checks

- **Missing** — variables in `.env.example` but not in your `.env`
- **Undocumented** — variables in your `.env` but not in `.env.example`
- **Unused** — variables defined in `.env` but never referenced in your code

Supports Node.js, Python, and Go projects automatically.

## Install
```bash
go install github.com/hyt4/envcheck@latest
```

## Usage
```bash
# run in current directory
envcheck

# specify custom file paths
envcheck --env .env.production --example .env.example

# ci mode — exits with code 1 if any issues found
envcheck --ci

# json output for scripting
envcheck --format json
```

## GitHub Actions example
```yaml
- name: Check .env
  run: envcheck --ci
```

## Built with

- [Cobra](https://github.com/spf13/cobra)
- [fatih/color](https://github.com/fatih/color)