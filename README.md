# ❄ Rui

Rui is my personal NixOS flake manager. It isn't very unique to my system at the
moment, so anyone can use it.

## Some Useful Commands

- `rui edit` - Open and edit your flake directory from anywhere
- `rui home/os switch` - Rebuild and switch your home or OS flake configuration
  from anywhere
- `rui home news` - Show the latest news from your Home Manager configuration
  packages

Rui checks the `FLAKE` environment variable for the path to your flake
directory.

Rui looks at the `FLAKE_EDITOR` environment variable for the editor to use when
opening the flake directory, but falls back to `EDITOR` if it isn't set.

## Installation

### Add to Flake Inputs

```nix
{
  inputs.rui = {
    url = "github:Fuwn/rui";
    inputs.nixpkgs.follows = "nixpkgs"; # Recommended
  };
}
```

### Add to System or Home Manager Packages

```nix
rui.packages.${pkgs.system}.default

# or

(import (
  pkgs.fetchFromGitHub {
    owner = "Fuwn";
    repo = "rui";
    rev = "...";  # Use the current commit revision hash
    hash = "..."; # Use the current commit sha256 hash
  }
)).packages.${builtins.currentSystem}.default
```

## `--help`

```text
NAME:
   rui - A new cli application

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
   --help, -h  show help

COPYRIGHT:
   Copyright (c) 2024-2024 Fuwn
```

## Licence

This project is licensed with the [GNU General Public License v3.0](./LICENSE.txt).
