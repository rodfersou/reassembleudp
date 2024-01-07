#!/bin/bash
LIGHT_PURPLE="\033[1;35m"
NO_COLOR="\033[0m"

COMMAND="brew services stop mongodb-community@7.0"

printf "${LIGHT_PURPLE}\$ "
echo $COMMAND
echo
printf $NO_COLOR

$COMMAND
