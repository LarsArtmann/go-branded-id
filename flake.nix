{
  description = "go-branded-id — Compile-time type-safe IDs for Go";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-parts = {
      url = "github:hercules-ci/flake-parts";
      inputs.nixpkgs-lib.follows = "nixpkgs";
    };
    systems.url = "github:nix-systems/default";
    treefmt-nix = {
      url = "github:numtide/treefmt-nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs =
    inputs@{
      self,
      nixpkgs,
      flake-parts,
      treefmt-nix,
      systems,
    }:
    flake-parts.lib.mkFlake { inherit inputs; } {
      systems = import systems;

      imports = [
        treefmt-nix.flakeModule
      ];

      perSystem =
        {
          config,
          lib,
          pkgs,
          system,
          ...
        }:
        let
          goPkg = pkgs.go_1_26;

          mkApp = name: runtimeInputs: text: {
            type = "app";
            program = "${pkgs.writeShellApplication { inherit name runtimeInputs text; }}/bin/${name}";
          };
        in
        {
          treefmt = {
            projectRootFile = "go.mod";
            programs = {
              gofumpt.enable = true;
              goimports.enable = true;
              golines.enable = true;
              nixfmt.enable = true;
            };
          };

          checks.format = config.treefmt.build.check self;
          devShells.default = pkgs.mkShellNoCC {
            packages = [
              goPkg
              pkgs.golangci-lint
              pkgs.gopls
              pkgs.trash-cli
            ];

            GOWORK = "off";
            GOEXPERIMENT = "jsonv2";

            shellHook = ''
              echo "go-branded-id dev shell — $(go version)"
            '';
          };

          devShells.ci = pkgs.mkShellNoCC {
            packages = [
              goPkg
              pkgs.golangci-lint
            ];

            GOWORK = "off";
            GOEXPERIMENT = "jsonv2";
          };

          checks = {
            build = pkgs.runCommand "go-branded-id-build" { nativeBuildInputs = [ goPkg ]; } ''
              export GOWORK=off
              export GOEXPERIMENT=jsonv2
              export GOCACHE="$TMPDIR/go-cache"
              cp -r ${
                lib.fileset.toSource {
                  root = ./.;
                  fileset = lib.fileset.gitTracked ./.;
                }
              } src && chmod -R u+w src && cd src
              ${goPkg}/bin/go build ./...
              touch $out
            '';
          };

          apps = {
            test = mkApp "test" [ goPkg ] ''
              export GOEXPERIMENT=jsonv2
              go test ./... -count=1 "$@"
            '';

            test-race = mkApp "test-race" [ goPkg ] ''
              export GOEXPERIMENT=jsonv2
              go test ./... -race -count=1 "$@"
            '';

            build = mkApp "build" [ goPkg ] ''
              export GOEXPERIMENT=jsonv2
              go build ./...
            '';

            vet = mkApp "vet" [ goPkg ] ''
              export GOEXPERIMENT=jsonv2
              go vet ./...
            '';

            lint = mkApp "lint" [ pkgs.golangci-lint ] ''
              export GOEXPERIMENT=jsonv2
              golangci-lint run ./...
            '';

            coverage = mkApp "coverage" [ goPkg ] ''
              export GOEXPERIMENT=jsonv2
              go test ./... -coverprofile=coverage.out -covermode=atomic "$@"
              go tool cover -func=coverage.out
            '';

            clean = mkApp "clean" [ goPkg pkgs.trash-cli ] ''
              trash-put coverage.out 2>/dev/null || true
              go clean -testcache
            '';
          };
        };
    };
}
