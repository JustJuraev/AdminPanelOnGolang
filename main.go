package main

import "net/http"

func Login(pahe http.ResponseWriter, r *http.Request) {

}

func main() {
	http.HandleFunc("/", Login)
	http.ListenAndServe("8081", nil)
}
