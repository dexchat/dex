{
  description = "DexChat IRC client";

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
            vendorHash = "sha256-H3O0g5AvWebQfPVleVtuSmKLfEd8KDpWt5Dmwxb23Ks=";

            subPackages = [ "cmd/tui" ];
            ldflags = [ "-s" "-w" ];
            env.CGO_ENABLED = 0;

            postInstall = ''
              mv "$out/bin/tui" "$out/bin/dex"
            '';

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
