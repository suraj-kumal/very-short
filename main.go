package very_short 

import (
	"fmt"
	"net/http"
)

func main(){

	//get cfg
	config := Load()
	
	//connect to database
	conn , err := DatabaseConnection(config.DatabaseURL)

	//if error is not absent
	if err != nil {
		log.Fatal(err)
	}
	
	//store connection reference to db 
	store := NewStore(conn)

	 




	http.ListenAndServe(": 8773", nil)
}
