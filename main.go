package main
import (
	"log"
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

	defer conn.Close()
	
	//store connection reference to db 
	store := NewStore(conn)

	h := New(store, config)

	mux := http.NewServeMux()

	h.RegisterRoutes(mux)

	log.Println("litenting on", config.PORT)

	log.Fatal(http.ListenAndServe(config.PORT, mux))


}
