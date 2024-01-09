#!/bin/bash
LIGHT_PURPLE="\033[1;35m"
NO_COLOR="\033[0m"

COMMAND="rabbitmq-server start -detach"

printf "${LIGHT_PURPLE}\$ "
echo $COMMAND
echo
printf $NO_COLOR

nix-shell -p rabbitmq-server --run "$COMMAND"
