{
  description = "Personal NixOS Flake Manager";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs";
    systems.url = "github:nix-systems/default";
    flake-compat.url = "https://flakehub.com/f/edolstra/flake-compat/1.tar.gz";

    flake-utils = {
      url = "github:numtide/flake-utils";
      inputs.systems.follows = "systems";
    };

    pre-commit-hooks = {
      url = "github:cachix/git-hooks.nix";
      inputs.nixpkgs.follows = "nixpkgs";
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
      in
      {
        packages.default = pkgs.buildGoModule {
          pname = "rui";
          version = "2024.09.19";
          src = pkgs.lib.cleanSource ./.;
          vendorHash = "sha256-mN/QjzJ4eGfbW1H92cCKvC0wDhCR6IUes2HCZ5YBdPA=";

          ldflags = [
            "-s"
            "-w"
          ];

          meta = with pkgs.lib; {
            description = "Personal NixOS Flake Manager";
            homepage = "https://github.com/Fuwn/rui";
            license = licenses.gpl3;
            maintainers = [ maintainers.Fuwn ];
            mainPackage = "rui";
            platforms = platforms.linux;
          };
        };

        apps.default = {
          type = "app";
          program = "${self.packages.${system}.default}/bin/rui";
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

          buildInputs = self.checks.${system}.pre-commit-check.enabledPackages;
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

                flake = mkOption {
                  type = types.str;
                  default = "";
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
