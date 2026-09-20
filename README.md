<h1 align="center">
  <img src="./resources/dexchat-icon.png" alt="Dex logo" width="128">
  <br>
  dexchat
</h1>
<p align="center">
  A modern and fast TUI IRC client with quick channel jumping, command palettes, and more.
  <br />
  <a href="#features">Features</a>
  ·
  <a href="#installation">Install</a>
  ·
  <a href="#configuration">Configuration</a>
  ·
  <a href="#contributing">Contributing</a>
</p>
<p align="center">
  <a href="https://github.com/dexchat/dex/releases"><img src="https://img.shields.io/github/v/release/dexchat/dex.svg" alt="Latest Release"></a>
  <a href="https://github.com/dexchat/dex/actions/workflows/release.yml"><img src="https://github.com/dexchat/dex/actions/workflows/release.yml/badge.svg" alt="Release Status"></a>
</p>

![preview](./resources/demo.gif)

## Features

- **Quick channel jumping:** move between channels quickly
- **Command palette:** access common actions from a searchable command menu
- **Fuzzy channel search:** Find channels without iteracting through all of them
- **Persistent history:** Preserve your conversations between sessions locally
- **Notifications:** Get notified about mentions and direct messages

## Installation
Use a package manager:
```bash
# Homebrew
brew install --cask dexchat/tap/dex

# Nix
nix profile install github:dexchat/dex
```

<details>
<summary><strong>Debian/Ubuntu</strong></summary>

```bash
curl -fsSL https://dexchat.org/apt/dexchat-keyring.gpg | sudo tee /usr/share/keyrings/dexchat-keyring.gpg >/dev/null
echo 'deb [signed-by=/usr/share/keyrings/dexchat-keyring.gpg] https://dexchat.org/apt stable main' | sudo tee /etc/apt/sources.list.d/dexchat.list >/dev/null
sudo apt update && sudo apt install dexchat
```
</details>

Until we add support for more package managers:
- [Packages](https://github.com/dexchat/dex/releases/latest) are available in Debian and RPM formats
- [Binaries](https://github.com/dexchat/dex/releases/latest) are available for Linux, macOS and Windows

You can also install it with Go:
```bash
go install github.com/dexchat/dex/cmd/tui@latest
```

Or, if you want test it before actually downloading dex, you can run its docker image and take a look at it:
```bash
docker run --rm -it ghcr.io/dexchat/dex:latest
```

## Configuration

`dexchat` reads its configuration from `$XDG_CONFIG_HOME/dex/config.toml`, or
`~/.config/dex/config.toml` (when `XDG_CONFIG_HOME` is not set). Start with the
provided example:

```bash
mkdir -p "$HOME/.config/dex"
wget -O "$HOME/.config/dex/config.toml" \
  https://raw.githubusercontent.com/dexchat/dex/main/config.example.toml
```

Edit the config file with your IRC server settings and preferences for the app. See
[`config.example.toml`](./config.example.toml) for all available options.

## Contributing
If you have any ideas, sugestions, issues or would like to contribute to dex, see the [contributing guide](https://github.com/dexchat/dex?tab=contributing-ov-file#contributing).
