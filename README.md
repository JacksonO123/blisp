# blisp

To run blisp

```sh
./run.sh [COMMAND] [FILENAME] [FLAGS]
```

[COMMAND] can be:
- `dev` => `go run *.go`
- `build` => `go build`
- `prod` => `./blisp`

Not including [FILENAME] [FLAGS] will start a repl

The `-b` flag will benchmark the evaluation

_Jackson Otto_
