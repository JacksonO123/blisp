ARGS="${@:2}"
if [[ "$1" == "dev" ]]; then
	go run *.go $ARGS
elif [[ "$1" == "prod" ]]; then
	./blisp $ARGS
elif [[ "$1" == "build" ]]; then
	go build
else
	echo "Command not recognized: \"$1\""
fi