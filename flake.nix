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
            vendorHash = "sha256-eswJCRF6jJLIgg5E8v04Yjt5tXWhhbbY497h0dsx+T8=";
          };
        };
      }
    );
}
