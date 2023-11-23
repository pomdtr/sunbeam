{
  description = "Wrap your CLIs in keyboard-friendly TUIs";

  inputs = {
    nixpkgs.url = "nixpkgs/nixos-unstable";
    utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, utils }: utils.lib.eachDefaultSystem
    (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        packages = {
          default = pkgs.buildGoModule {
            name = "sunbeam";
            src = self;
            vendorHash = "sha256-HbAgvGp375KxNwyy5cBv19IoTyHwCz7S+4Nk0osmx8A=";
          };
        };
      }
    );
}
