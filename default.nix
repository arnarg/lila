{
  lib,
  buildGoApplication,
  nvd,
  makeWrapper,
}:
buildGoApplication {
  name = "lila";

  src = lib.cleanSource ./.;

  modules = ./gomod2nix.toml;

  subPackages = ["cmd/lila"];

  nativeBuildInputs = [makeWrapper];

  postInstall = ''
    wrapProgram $out/bin/lila --prefix PATH : ${lib.makeBinPath [nvd]}
  '';
}
