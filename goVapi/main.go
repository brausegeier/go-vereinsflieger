package main

import (
	"fmt"
	_ "github.com/janvogt/go-vereinsflieger/handler"
	"net/http"
)

func main() {
	err := http.ListenAndServe(":8080", nil)
	fmt.Println(err)
}
