## Initial Bootstrap

1. Install nix Follow the instructions according your OS
   [here](https://nixos.org/download.html)

2. Install direnv

```bash
nix-env -i direnv
```

3. Configure direnv by adding this in your ~/.zshrc

```
eval "$(direnv hook zsh)"
```

4. Copy .env example configuration

```bash
cp docs/dotenv.example .env
```

5. Allow direnv open the shell in the project

```bash
direnv allow
```
