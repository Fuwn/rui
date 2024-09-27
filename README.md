# ❄ Rui

Rui is my personal NixOS flake manager. It isn't very unique to my system at the
moment, so anyone can use it.

## Useful Commands

- `rui edit` - Open and edit your flake directory from anywhere
- `rui home/os switch` - Rebuild and switch your home or OS flake configuration
  from anywhere
- `rui home news` - Show the latest news from your Home Manager configuration
  packages

## Installation

### Add to Flake Inputs (for Flakes Users)

```nix
{
  inputs.rui = {
    url = "github:Fuwn/rui";
    inputs.nixpkgs.follows = "nixpkgs"; # Recommended
  };
}
```

### Add to Home Manager (Managed Configuration)

This method manages the configuration for you with Nix.

```nix
# ...

inputs.home-manager.lib.homeManagerConfiguration {
  modules = [
    inputs.rui.homeManagerModules.${builtins.currentSystem}.default
  ];
};

# ...
```

### Configure Rui Using Home Manager

```nix
{
  programs.rui = {
    enable = true; # Defaults to false

    settings = {
      # Status notifications via `notify-send`; defaults to false
      notify = true;

      # The command to use for sending notifications, view a cool example below;
      # defaults to `notify-send`
      notifier = "notify-send";

      # Rui falls back on the `FLAKE_EDITOR` and `EDITOR` environment variables
      # if `editor` is unset
      editor = "code";

      # Rui falls back on the `FLAKE` environment variable if `flake` is unset
      flake = "/path/to/your-flake";

      # Allow unfree packages; defaults to false
      allow-unfree = false;

      # Extra arguments to pass to `nixos-rebuild` and `home-manager`; defaults
      # to [ ]
      extra-args = [ "--impure" ];
    };
  };
}
```

### Add to System or Home Manager Packages (Manual Configuration)

Using this method, configuration is done manually by the user in the
`$HOME/.config/rui/config.json` file.

```nix
# For flakes users
rui.packages.${pkgs.system}.default

# For non-flakes users
(import (
  pkgs.fetchFromGitHub {
    owner = "Fuwn";
    repo = "rui";
    rev = "...";  # Use the current commit revision hash
    hash = "..."; # Use the current commit sha256 hash
  }
)).packages.${builtins.currentSystem}.default
```

## Custom Notification Command Example

Rui uses `notify-send` by default for sending notifications, but you can set
the `notifier` configuration value to any file path. Here's an example of a
distributed notification script that sends notifications to your phone **and**
your PC. This can easily be adapted to send notifications to any service, e.g.,
Telegram, Discord, other webhook receivers, etc.

This example uses [Bark](https://bark.day.app/#/?id=%E6%BA%90%E7%A0%81), an
extremely simple and easy-to-use notification service for iOS devices.

```sh
#!/usr/bin/env dash

# Send a notification to your host PC
notify-send "$1" "$2"

# Send a notification to your iOS device
curl -X "POST" "https://api.day.app/your_bark_api_key" \
  -H 'Content-Type: application/json; charset=utf-8' \
  --silent \
  -d '{
    "body": "'"${2}"'",
    "title": "'"${1}"'",
    "icon": "https://nixos.wiki/images/thumb/2/20/Home-nixos-logo.png/207px-Home-nixos-logo.png"
  }'
```

## `--help`

```text
NAME:
   rui - Personal NixOS Flake Manager

USAGE:
   rui [global options] command [command options]

DESCRIPTION:
   Personal NixOS Flake Manager

AUTHOR:
   Fuwn <contact@fuwn.me>

COMMANDS:
   home
   os
   edit
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --allow-unfree  (default: false)
   --help, -h      show help

COPYRIGHT:
   Copyright (c) 2024-2024 Fuwn
```

## Licence

This project is licensed with the [GNU General Public License v3.0](./LICENSE.txt).
