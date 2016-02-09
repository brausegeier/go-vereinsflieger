package main

import (
	"fmt"
	"github.com/janvogt/go-vereinsflieger/handler"
	"net/http"
)

func main() {
	err := http.ListenAndServe(fmt.Sprintf(":%d", handler.DefaultConfig.Port), nil)
	fmt.Println(err)
}
