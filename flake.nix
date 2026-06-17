{
  description = "UPD — Upgrade NPM Package Dependencies (Go port)";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-parts.url = "github:hercules-ci/flake-parts";
  };

  outputs =
    inputs@{ self, nixpkgs, flake-parts, ... }:
    flake-parts.lib.mkFlake { inherit inputs; } {
      systems = [
        "x86_64-linux"
        "aarch64-linux"
        "x86_64-darwin"
        "aarch64-darwin"
      ];

      perSystem =
        { pkgs, ... }:
        let
          version = "1.0.0";
          buildFlags = "-ldflags=-s -w -trimpath";
        in
        {
          packages = {
            default = pkgs.buildGoModule {
              pname = "upd";
              inherit version;
              src = ./.;
              vendorHash = null;
              subPackages = [ "cmd/upd" ];
              ldflags = [ "-s" "-w" ];
              meta = with pkgs.lib; {
                description = "Upgrade NPM package dependencies while preserving formatting";
                homepage = "https://github.com/LarsArtmann/upd";
                license = licenses.mit;
                mainProgram = "upd";
              };
            };
          };

          devShells.default = pkgs.mkShell {
            buildInputs = with pkgs; [
              go
              gopls
              gotools
              go-tools # staticcheck
            ];
          };

          apps = {
            build = {
              type = "app";
              program = pkgs.writeShellApplication {
                name = "build";
                runtimeInputs = [ pkgs.go ];
                text = ''
                  go build ${buildFlags} -o bin/upd ./cmd/upd
                '';
              };
            };

            test = {
              type = "app";
              program = pkgs.writeShellApplication {
                name = "test";
                runtimeInputs = [ pkgs.go ];
                text = ''
                  go test ./... -v -count=1
                '';
              };
            };

            lint = {
              type = "app";
              program = pkgs.writeShellApplication {
                name = "lint";
                runtimeInputs = with pkgs; [ go gopls ];
                text = ''
                  go vet ./... && echo "vet OK"
                  go build ./... && echo "build OK"
                '';
              };
            };

            run = {
              type = "app";
              program = pkgs.writeShellApplication {
                name = "run";
                runtimeInputs = [ pkgs.go ];
                text = ''
                  go run ./cmd/upd "$@"
                '';
              };
            };
          };
        };
    };
}
