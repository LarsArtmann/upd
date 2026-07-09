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
          goExperiment = "jsonv2";
        in
        {
          formatter = pkgs.nixpkgs-fmt;

          packages = {
            default = pkgs.buildGoModule {
              pname = "upd";
              inherit version;
              src = ./.;
              vendorHash = "sha256-HHBnbQrRKhy4EGNZfFyo8C7qHzhAASITutgYa4eHADU=";
              subPackages = [ "cmd/upd" ];
              env.GOEXPERIMENT = goExperiment;
              ldflags = [
                "-s"
                "-w"
                "-X"
                "github.com/LarsArtmann/upd.ProgramVersion=${version}"
              ];
              meta = with pkgs.lib; {
                description = "Upgrade NPM package dependencies while preserving formatting";
                homepage = "https://github.com/LarsArtmann/upd";
                license = licenses.mit;
                mainProgram = "upd";
              };
            };
          };

          devShells.default = pkgs.mkShell {
            GOEXPERIMENT = goExperiment;
            buildInputs = with pkgs; [
              go
              gopls
              gotools
              go-tools # staticcheck
              golangci-lint
              govulncheck
              vhs
              ttyd
              ffmpeg
            ];
          };

          apps = {
            build = {
              type = "app";
              program = pkgs.writeShellApplication {
                name = "build";
                runtimeInputs = [ pkgs.go ];
                text = ''
                  export GOEXPERIMENT=${goExperiment}
                  go build -trimpath -ldflags='-s -w -X github.com/LarsArtmann/upd.ProgramVersion=${version}' -o bin/upd ./cmd/upd
                '';
              };
              meta.description = "Build upd to bin/upd";
            };

            test = {
              type = "app";
              program = pkgs.writeShellApplication {
                name = "test";
                runtimeInputs = [ pkgs.go ];
                text = ''
                  export GOEXPERIMENT=${goExperiment}
                  go test ./... -v -count=1
                '';
              };
              meta.description = "Run all tests with verbose output";
            };

            lint = {
              type = "app";
              program = pkgs.writeShellApplication {
                name = "lint";
                runtimeInputs = with pkgs; [ go golangci-lint ];
                text = ''
                  export GOEXPERIMENT=${goExperiment}
                  go vet ./... && echo "vet OK"
                  go build ./... && echo "build OK"
                  golangci-lint run ./... && echo "lint OK"
                '';
              };
              meta.description = "Run go vet, build check, and golangci-lint";
            };

            run = {
              type = "app";
              program = pkgs.writeShellApplication {
                name = "run";
                runtimeInputs = [ pkgs.go ];
                text = ''
                  export GOEXPERIMENT=${goExperiment}
                  go run ./cmd/upd "$@"
                '';
              };
              meta.description = "Run upd from source with arguments";
            };

            demo = {
              type = "app";
              program = pkgs.writeShellApplication {
                name = "demo";
                runtimeInputs = with pkgs; [ go vhs ttyd ffmpeg git ];
                text = ''
                  export GOEXPERIMENT=${goExperiment}
                  build_dir="$(mktemp -d)"
                  trap 'rm -rf "$build_dir"' EXIT

                  repo_root="$(git rev-parse --show-toplevel)"
                  go build -C "$repo_root" -trimpath \
                    -ldflags='-s -w -X github.com/LarsArtmann/upd.ProgramVersion=${version}' \
                    -o "$build_dir/upd" ./cmd/upd

                  export PATH="$build_dir:$PATH"
                  cd "$repo_root/demo"

                  if [ "$#" -gt 0 ] && [ "$1" = "--publish" ]; then
                    shift
                    for tape in *.tape; do
                      echo "Rendering and publishing $tape..."
                      vhs --publish "$tape"
                    done
                  else
                    for tape in *.tape; do
                      echo "Rendering $tape (local only)..."
                      vhs "$tape"
                    done
                    echo ""
                    echo "Done. GIFs are in demo/."
                    echo "To publish to vhs.charm.sh: nix run .#demo -- --publish"
                  fi
                '';
              };
              meta.description = "Render VHS demo GIFs locally or publish to vhs.charm.sh";
            };
          };
        };
    };
}
