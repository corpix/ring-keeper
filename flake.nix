{
  inputs = {
    nixpkgs.url = "tarball+https://git.tatikoma.dev/corpix/nixpkgs/archive/d5b349c0af56d2fa34ce149895fd35df9c54c844.tar.gz";
  };

  outputs = { self, nixpkgs }:
    let
      arch = "x86_64-linux";
      pkgs = nixpkgs.legacyPackages.${arch}.pkgs;

      inherit (pkgs)
        writeScript
        stdenv
        mkShell
      ;

      inherit (pkgs)
        go
        git
        gnumake
      ;
    in rec {
      devShells.${arch}.default = mkShell {
        name = "ring-keeper";
        packages = [
          go
          git
          gnumake
        ];
      };
    };
}
