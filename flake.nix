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
            vendorSha256 = "sha256-fwXNoxJMg0g7koNDvGgMXFjbbsSHdl0K/or7OzIznrk=";
          };
        };
      }
    );
}
