with import <nixpkgs> {};
pkgs.mkShell {
    name = "go-shell";
    packages = [
        (python311.withPackages(ps: with ps; [
            pymongo
            pytest
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
        tmux
    ];

    shellHook = ''
        echo "Starting Go development environment"

        export GOPATH=$ENV_DIR/.cache/go
        mkdir -p $GOPATH
        export MONGOPATH=$ENV_DIR/.cache/mongo
        mkdir -p $MONGOPATH
        export PATH="$GOPATH/bin:$PATH"

        go mod download
    '';
}
