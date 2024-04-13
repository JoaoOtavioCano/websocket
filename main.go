package main

import (
	"fmt"
	"log"
	"net/http"
	"github.com/JoaoOtavioCano/websocket"
)

func main() {

	runServer()

}

func runServer() {
	http.HandleFunc("/", func (w http.ResponseWriter, r *http.Request) {
		fmt.Println(r.Method)	
		fmt.Println(r.Header)	
	})


	log.Println("Server live")
	log.Println(http.ListenAndServe(":8000", nil))
}
