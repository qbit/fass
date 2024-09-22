{
  description = "fass: stuff and fass";

  inputs.nixpkgs.url = "nixpkgs/nixos-unstable";

  outputs =
    { self
    , nixpkgs
    ,
    }:
    let
      supportedSystems = [ "x86_64-linux" "x86_64-darwin" "aarch64-linux" "aarch64-darwin" ];
      forAllSystems = nixpkgs.lib.genAttrs supportedSystems;
      nixpkgsFor = forAllSystems (system: import nixpkgs { inherit system; });
    in
    {
      overlay = _: prev: { inherit (self.packages.${prev.system}) fass; };

      packages = forAllSystems (system:
        let
          pkgs = nixpkgsFor.${system};
        in
        {
          fass = with pkgs; pkgs.buildGoModule rec {
            pname = "fass";
            version = "v0.0.0";
            src = ./.;

            vendorHash = "sha256-gyY6JQn2r23Gwh0Dop5BoBT7TbZIxM2Ed+sC+yqTzDQ=";

            nativeBuildInputs = [ pkg-config copyDesktopItems ];
            buildInputs = [
              fyne
              glfw
              libGL
              libGLU
              openssh
              pkg-config
              glibc
              xorg.libXcursor
              xorg.libXi
              xorg.libXinerama
              xorg.libXrandr
              xorg.libXxf86vm
              xorg.xinput
            ];

            buildPhase = ''
              ${fyne}/bin/fyne package
            '';

            installPhase = ''
                mkdir -p $out
                pkg="$PWD/${pname}.tar.xz"
                cd $out
                tar --strip-components=1 -xvf $pkg
             '';
          };
        });

      defaultPackage = forAllSystems (system: self.packages.${system}.fass);
      devShells = forAllSystems (system:
        let
          pkgs = nixpkgsFor.${system};
        in
        {
          default = pkgs.mkShell {
            shellHook = ''
              PS1='\u@\h:\@; '
              nix run github:qbit/xin#flake-warn
              echo "Go `${pkgs.go}/bin/go version`"
            '';
            buildInputs = with pkgs; [
              fyne
              git
              go
              gopls
              go-tools
              glxinfo

              glfw
              glibc
              pkg-config
              xorg.libXcursor
              xorg.libXi
              xorg.libXinerama
              xorg.libXrandr
              xorg.libXxf86vm
              xorg.xinput
              graphviz

              go-font
            ];
          };
        });
    };
}
