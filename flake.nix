{
  description = "Wrap your CLIs in keyboard-friendly TUIs";

  inputs = {
    nixpkgs.url = "nixpkgs/nixos-23.05";
  };

  outputs = { nixpkgs, ... }:
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
            vendorSha256 = "sha256-3gMP9VjC8+u41gvzT45LflqZ4uk5+tObBtlJO5SCjwQ=";
          };
        };
    };
}
