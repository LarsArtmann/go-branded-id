{
  description = "go-branded-id — Compile-time type-safe IDs for Go";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-parts = {
      url = "github:hercules-ci/flake-parts";
      inputs.nixpkgs-lib.follows = "nixpkgs";
    };
    treefmt-nix = {
      url = "github:numtide/treefmt-nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs =
    inputs@{
      self,
      flake-parts,
      treefmt-nix,
      ...
    }:
    flake-parts.lib.mkFlake { inherit inputs; } {
      # nixpkgs unstable (26.11) dropped x86_64-darwin; aarch64-darwin remains supported.
      systems = [
        "x86_64-linux"
        "aarch64-linux"
        "aarch64-darwin"
      ];

      imports = [
        treefmt-nix.flakeModule
      ];

      perSystem =
        {
          config,
          lib,
          pkgs,
          ...
        }:
        let
          # Root module floor: go.mod says `go 1.27.0` — the consumer-facing
          # minimum. Root checks build with exactly this toolchain as tripwire.
          goPkg = pkgs.go_1_27;

          # Linter module floor: linter/go.mod says `go 1.27.1` (it depends on
          # go-finding, which requires 1.27.1).
          linterGoPkg = pkgs.go_1_27;

          # The brandid-lint CLI: hermetic build + tests via buildGoModule
          # (the module has dependencies, so a raw runCommand sandbox fails on
          # the module cache). Overridden to go_1_27 because the nixpkgs
          # default Go (1.26) is below the linter's `go 1.27.1` floor.
          # Version tracks linter/CHANGELOG.md (tags `linter/vX.Y.Z`) — bump
          # it in the same change as a linter release.
          brandidLint = pkgs.buildGoModule.override { go = linterGoPkg; } {
            pname = "brandid-lint";
            version = "0.5.1";
            src = lib.fileset.toSource {
              root = ./linter;
              fileset = lib.fileset.gitTracked ./linter;
            };
            # Updated via `buildflow -s nix-hash-fix --fix` on FOD mismatch.
            vendorHash = "sha256-4w3VE1dxWjOsE/MuJWfQBsMxPs+M/PdTNDRanmh6PU4=";
            doCheck = true;
          };

          mkApp = name: runtimeInputs: text: {
            type = "app";
            program = "${
              pkgs.writeShellApplication {
                inherit name runtimeInputs text;
                # Per-module toolchain dirs + explicit mode pins: both floors
                # are go1.27, where the goexperiment.jsonv2 tag is default-true
                # — so every "v1" leg pins GOEXPERIMENT=nojsonv2 and every v2
                # leg jsonv2. Without the pins an ambient GOEXPERIMENT (or the
                # go1.27 default) silently turns a v1 leg into a second v2 leg.
                runtimeEnv = {
                  rootGoBin = "${goPkg}/bin";
                  linterGoBin = "${linterGoPkg}/bin";
                };
              }
            }/bin/${name}";
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
            # `go` resolves to the shared 1.27 floor of both modules (1.27.0
            # root, 1.27 linter) so gopls works in either directory. The
            # checks below pin the true floors.
            packages = [
              linterGoPkg
              pkgs.golangci-lint
              pkgs.gopls
              pkgs.trash-cli
            ];

            GOWORK = "off";

            shellHook = ''
              echo "go-branded-id dev shell — $(go version)"
            '';
          };

          devShells.ci = pkgs.mkShellNoCC {
            packages = [
              goPkg
              linterGoPkg
              pkgs.golangci-lint
            ];

            GOWORK = "off";
          };

          checks = {
            build = pkgs.runCommand "go-branded-id-build" { nativeBuildInputs = [ goPkg ]; } ''
              export GOWORK=off
              export GOCACHE="$TMPDIR/go-cache"
              cp -r ${
                lib.fileset.toSource {
                  root = ./.;
                  fileset = lib.fileset.gitTracked ./.;
                }
              } src && chmod -R u+w src && cd src
              GOEXPERIMENT=nojsonv2 ${goPkg}/bin/go build ./...
              GOEXPERIMENT=jsonv2 ${goPkg}/bin/go build ./...
              touch $out
            '';

            test = pkgs.runCommand "go-branded-id-test" { nativeBuildInputs = [ goPkg ]; } ''
              export GOWORK=off
              export GOCACHE="$TMPDIR/go-cache"
              cp -r ${
                lib.fileset.toSource {
                  root = ./.;
                  fileset = lib.fileset.gitTracked ./.;
                }
              } src && chmod -R u+w src && cd src
              GOEXPERIMENT=nojsonv2 ${goPkg}/bin/go test ./... -count=1
              GOEXPERIMENT=jsonv2 ${goPkg}/bin/go test ./... -count=1
              touch $out
            '';

            linter = brandidLint;
          };

          packages = {
            brandid-lint = brandidLint;
            default = brandidLint;
          };

          apps = {
            # nix run .#brandid-lint -- -rules BD003 .
            brandid-lint = {
              type = "app";
              program = "${brandidLint}/bin/brandid-lint";
            };

            test = mkApp "test" [ ] ''
              echo "=== Testing json v1 (go1.27) ==="
              PATH="$rootGoBin:$PATH" GOEXPERIMENT=nojsonv2 go test ./... -count=1 "$@"
              echo "=== Testing json v2 (go1.27) ==="
              PATH="$rootGoBin:$PATH" GOEXPERIMENT=jsonv2 go test ./... -count=1 "$@"
              echo "=== Testing brandid-lint module (go1.27) ==="
              (cd linter && PATH="$linterGoBin:$PATH" go test ./... -count=1 "$@")
            '';

            test-race = mkApp "test-race" [ ] ''
              echo "=== Race testing json v1 (go1.27) ==="
              PATH="$rootGoBin:$PATH" GOEXPERIMENT=nojsonv2 go test ./... -race -count=1 "$@"
              echo "=== Race testing json v2 (go1.27) ==="
              PATH="$rootGoBin:$PATH" GOEXPERIMENT=jsonv2 go test ./... -race -count=1 "$@"
              echo "=== Race testing brandid-lint module (go1.27) ==="
              (cd linter && PATH="$linterGoBin:$PATH" go test ./... -race -count=1 "$@")
            '';

            build = mkApp "build" [ ] ''
              echo "=== Building json v1 (go1.27) ==="
              PATH="$rootGoBin:$PATH" GOEXPERIMENT=nojsonv2 go build ./...
              echo "=== Building json v2 (go1.27) ==="
              PATH="$rootGoBin:$PATH" GOEXPERIMENT=jsonv2 go build ./...
              echo "=== Building brandid-lint module (go1.27) ==="
              (cd linter && PATH="$linterGoBin:$PATH" go build ./...)
            '';

            vet = mkApp "vet" [ ] ''
              PATH="$rootGoBin:$PATH" GOEXPERIMENT=nojsonv2 go vet ./...
              (cd linter && PATH="$linterGoBin:$PATH" go vet ./...)
            '';

            lint = mkApp "lint" [ pkgs.golangci-lint ] ''
              echo "=== Linting json v1 (go1.27) ==="
              PATH="$rootGoBin:$PATH" GOEXPERIMENT=nojsonv2 golangci-lint run ./...
              echo "=== Linting json v2 (go1.27) ==="
              PATH="$rootGoBin:$PATH" golangci-lint run --build-tags goexperiment.jsonv2 ./...
              echo "=== Linting brandid-lint module (go1.27) ==="
              (cd linter && PATH="$linterGoBin:$PATH" golangci-lint run ./...)
            '';

            coverage = mkApp "coverage" [ ] ''
              PATH="$rootGoBin:$PATH" GOEXPERIMENT=nojsonv2 go test ./... -coverprofile=coverage.out -covermode=atomic "$@"
              PATH="$rootGoBin:$PATH" go tool cover -func=coverage.out
              (cd linter \
                && PATH="$linterGoBin:$PATH" go test ./... -coverprofile="$TMPDIR/linter-coverage.out" -covermode=atomic "$@" \
                && PATH="$linterGoBin:$PATH" go tool cover -func="$TMPDIR/linter-coverage.out")
            '';

            clean = mkApp "clean" [ pkgs.trash-cli ] ''
              trash-put coverage.out 2>/dev/null || true
              PATH="$rootGoBin:$PATH" go clean -testcache
            '';
          };
        };
    };
}
