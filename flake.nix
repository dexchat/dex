{
  description = "dexchat IRC client";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

  outputs = { nixpkgs, ... }:
    let
      supportedSystems = [
        "x86_64-linux"
        "aarch64-linux"
        "aarch64-darwin"
      ];
      forAllSystems = nixpkgs.lib.genAttrs supportedSystems;
    in
    {
      packages = forAllSystems (system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
        in
        {
          default = pkgs.buildGoModule {
            pname = "dexchat";
            version = "0.1.1-dev";

            src = ./.;
            vendorHash = "sha256-lor5WtPMd0ea09ICchAv648K8DzUg2aouTizFZbbpnM=";

            subPackages = [ "cmd/dex" ];
            ldflags = [ "-s" "-w" ];
            env.CGO_ENABLED = 0;

            meta = {
              description = "A modern and fast terminal IRC client";
              homepage = "https://dexchat.org";
              license = nixpkgs.lib.licenses.mit;
              mainProgram = "dex";
            };
          };
        });

      devShells = forAllSystems (system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
        in
        {
          default = pkgs.mkShell {
            packages = [ pkgs.go ];
          };
        });
    };
}
