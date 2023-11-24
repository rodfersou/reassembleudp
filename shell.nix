with import <nixpkgs> {};
pkgs.mkShell {
    name = "go-shell";
    packages = [
        (python311.withPackages(ps: with ps; [
            supervisor
            pymongo
        ]))
        go_1_20
        nodejs-19_x
        mongodb-6_0
        lazygit
        ripgrep
        tree
        pre-commit
        delve
        entr
    ];

    shellHook = ''
        echo "Starting Go development environment"

        export GOPATH=$ENV_DIR/.direnv
        export PATH="$GOPATH/bin:$ENV_DIR/node_modules/.bin:$PATH"

        go mod download
    '';
}
