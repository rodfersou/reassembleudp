#!/bin/bash
LIGHT_PURPLE="\033[1;35m"
NO_COLOR="\033[0m"

rm -rf inbox/*
COMMAND="go run reassembleudp.go"

printf "${LIGHT_PURPLE}\$ "
echo $COMMAND
echo
printf $NO_COLOR
$COMMAND
