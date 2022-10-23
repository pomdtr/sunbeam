{ pkgs }: {
    deps = [
        pkgs.go_1_18
        pkgs.gopls
        pkgs.nodePackages.zx
    ];
}