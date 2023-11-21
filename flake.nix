{
  description = "Wrap your CLIs in keyboard-friendly TUIs";

  inputs = {
    nixpkgs.url = "nixpkgs/nixos-unstable";
  };

  outputs = { self, nixpkgs, ... }:
    let
      # you can also put any architecture you want to support here
      # i.e. aarch64-darwin for never M1/2 macbooks
      system = "aarch64-darwin";
      pname = "sunbeam";
    in
    {
      packages.${system} =
        let
          pkgs = nixpkgs.legacyPackages.${system}; # this gives us access to nixpkgs as we are used to
        in
        {
          default = pkgs.buildGoModule {
            name = pname;
            src = self;
            vendorHash = "sha256-HbAgvGp375KxNwyy5cBv19IoTyHwCz7S+4Nk0osmx8A=";
          };
        };
    };
}
