package very_short 

import (
	"fmt"
	"net/http"
)

func main(){
	config := Load()

	db , err := DatabaseConnection(config.DatabaseURL)

	if err != nil {
		log.Fatal(err)
	}


	http.ListenAndServe(": 8773", nil)
}
