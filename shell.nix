with import <nixpkgs> {};
pkgs.mkShell {
    name = "go-shell";
    packages = [
        (python311.withPackages(ps: with ps; [
            pymongo
            pytest
        ]))
        delve
        entr
        go_1_20
        lazygit
        nodejs-18_x
        pre-commit
        ripgrep
        tmux
        tree
    ];

    shellHook = ''
        echo "Starting Go development environment"

        export GOPATH=$ENV_DIR/.cache/go
        mkdir -p $GOPATH
        export MONGOPATH=$ENV_DIR/.cache/mongo
        mkdir -p $MONGOPATH
        export PATH="$GOPATH/bin:$PATH"

        go mod download
        [ ! -f .env ] && cp docs/dotenv.example .env
    '';
}
