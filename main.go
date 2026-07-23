package main 

import (
	"fmt"
	"net/http"
)

func main(){
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request){
		fmt.Fprint(w, "Hello, You have requested : %s\n", r.URL.Path)

	})
	http.ListenAndServe(": 80", nil)
}
