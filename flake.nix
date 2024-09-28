{
  description = "Personal NixOS Flake Manager";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs";
    systems.url = "github:nix-systems/default";

    flake-compat = {
      url = "github:edolstra/flake-compat";
      flake = false;
    };

    flake-utils = {
      url = "github:numtide/flake-utils";
      inputs.systems.follows = "systems";
    };

    pre-commit-hooks = {
      url = "github:cachix/git-hooks.nix";

      inputs = {
        flake-compat.follows = "flake-compat";
        nixpkgs.follows = "nixpkgs";
      };
    };
  };

  outputs =
    {
      self,
      nixpkgs,
      flake-utils,
      pre-commit-hooks,
      ...
    }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = import nixpkgs { inherit system; };

        meta = with pkgs.lib; {
          description = "Personal NixOS Flake Manager";
          homepage = "https://github.com/Fuwn/rui";
          license = licenses.gpl3;
          maintainers = [ maintainers.Fuwn ];
          mainPackage = "rui";
          platforms = platforms.linux;
        };

        rui = pkgs.buildGoModule {
          inherit meta;

          pname = "rui";
          version = "2024.09.27";
          src = pkgs.lib.cleanSource ./.;
          vendorHash = "sha256-mN/QjzJ4eGfbW1H92cCKvC0wDhCR6IUes2HCZ5YBdPA=";

          ldflags = [
            "-s"
            "-w"
          ];
        };
      in
      {
        packages = {
          default = rui;
          rui = self.packages.${system}.default;
        };

        apps = {
          default = {
            inherit meta;

            type = "app";
            program = "${self.packages.${system}.default}/bin/rui";
          };

          rui = self.apps.${system}.default;
        };

        formatter = nixpkgs.legacyPackages."${system}".nixfmt-rfc-style;

        checks.pre-commit-check = pre-commit-hooks.lib.${system}.run {
          src = ./.;

          hooks = {
            deadnix.enable = true;
            flake-checker.enable = true;
            nixfmt-rfc-style.enable = true;
            statix.enable = true;
          };
        };

        devShells.default = nixpkgs.legacyPackages.${system}.mkShell {
          inherit (self.checks.${system}.pre-commit-check) shellHook;

          buildInputs = self.checks.${system}.pre-commit-check.enabledPackages ++ [
            pkgs.go_1_22
          ];
        };

        homeManagerModules.default =
          { config, ... }:
          with pkgs.lib;
          {
            options.programs.rui = {
              enable = mkOption {
                type = types.bool;
                default = false;
              };

              settings = {
                editor = mkOption {
                  type = types.str;
                  default = "";
                };

                notify = mkOption {
                  type = types.bool;
                  default = false;
                };

                notifier = mkOption {
                  type = types.str;
                  default = "notify-send";
                };

                flake = mkOption {
                  type = types.str;
                  default = "";
                };

                allow-unfree = mkOption {
                  type = types.bool;
                  default = false;
                };

                extra-args = mkOption {
                  type = types.listOf types.str;
                  default = [ ];
                };
              };
            };

            config = mkIf config.programs.rui.enable {
              home.packages = [
                self.packages.${system}.default
                pkgs.libnotify
              ];

              xdg.configFile."rui/config.json".text = builtins.toJSON config.programs.rui.settings;
            };
          };
      }
    );
}
