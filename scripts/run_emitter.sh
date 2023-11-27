#!/bin/bash
LIGHT_PURPLE="\033[1;35m"
NO_COLOR="\033[0m"

COMMAND="node udp_emitter.js"

printf "${LIGHT_PURPLE}\$ "
echo $COMMAND
echo
printf $NO_COLOR
$COMMAND
