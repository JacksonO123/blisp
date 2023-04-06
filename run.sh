#!/bin/bash
args="${*:2}"
if [[ "$1" == "dev" ]]; then
	go run ./*.go $args
elif [[ "$1" == "prod" ]]; then
	./blisp $args
elif [[ "$1" == "build" ]]; then
	echo b
	go build
else
	echo "Command not recognized: \"$1\""
fi
