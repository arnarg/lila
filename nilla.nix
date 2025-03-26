let
  src = import ./npins;

  nilla = import src.nilla;

  systems = ["x86_64-linux" "aarch64-linux"];
in
  nilla.create ({config, ...}: {
    config = {
      inputs = {
        nixpkgs = {
          src = src.nixpkgs;

          settings.overlays = [
            (import "${src.gomod2nix}/overlay.nix")
          ];
        };
      };

      packages.default = {
        inherit systems;

        builder = "nixpkgs";
        settings.pkgs = config.inputs.nixpkgs.result;

        package = import ./default.nix;
      };

      shells.default = {
        inherit systems;

        builder = "nixpkgs";
        settings.pkgs = config.inputs.nixpkgs.result;

        shell = {
          mkShellNoCC,
          npins,
          gomod2nix,
          nvd,
          ...
        }:
          mkShellNoCC {
            packages = [
              npins
              gomod2nix
              nvd
            ];
          };
      };
    };
  })
