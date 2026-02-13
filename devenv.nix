{pkgs, config, lib, ...}: {
  languages.go = {
    enable = true;
  };

  packages = with pkgs; [
    golangci-lint
  ];

  scripts = { 
    lint = {
      exec = ''
        golangci-lint run
      '';
      description = "Run golangci-lint to check for code issues.";
    };
    lint-fix = {
      exec = ''
        golangci-lint run --fix
      '';
      description = "Run golangci-lint with --fix to automatically fix issues.";
    };

    test-unit = {
      exec = ''
        go test -v -race -timeout=1m -covermode=atomic ./...
      '';
      description = "Run unit tests for all packages.";
    };
    test-integration = {
      exec = ''
         go test -v -race -timeout=1m -tags=integration ./test/...
      '';
      description = "Run integration tests.";
    };
  };

  enterShell = ''
    echo
    echo "Available scripts:"
    ${pkgs.gnused}/bin/sed -e 's| |••|g' -e 's|=| |' <<EOF | ${pkgs.util-linuxMinimal}/bin/column -t | ${pkgs.gnused}/bin/sed -e 's|^|  |' -e 's|••| |g'
    ${lib.generators.toKeyValue {} (lib.mapAttrs (name: value: value.description) config.scripts)}
    EOF
    echo
  '';
}
