package main

import (
	"fmt"
	"github.com/brausegeier/go-vereinsflieger/api"
	"net/http"
)

func main() {
	http.Handle("/voucher", api.FailableHandler(api.AddVoucher))
	fmt.Printf("Listening at :%d...", api.DefaultConfig.Port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", api.DefaultConfig.Port), nil)
	fmt.Println(err)
}
