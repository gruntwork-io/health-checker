package server

import (
	"fmt"
	"net/http"
	"github.com/urfave/cli"
)

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

func StartHttpServer(cliContext *cli.Context) {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}
