{
  lib,
  buildGoApplication,
  nvd,
  makeWrapper,
}: let
  version = "0.0.0";
in
  buildGoApplication {
    inherit version;
    pname = "lila";

    src = lib.cleanSource ./.;

    modules = ./gomod2nix.toml;

    subPackages = ["cmd/lila"];
    ldflags = ["-X main.version=${version}"];

    nativeBuildInputs = [makeWrapper];

    postInstall = ''
      wrapProgram $out/bin/lila --prefix PATH : ${lib.makeBinPath [nvd]}
    '';
  }
