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


	cache := NewCache(1000)

	h := New(store, config, cache)

	mux := http.NewServeMux()

	h.RegisterRoutes(mux)

	log.Println("listening on :", config.PORT)

	log.Fatal(http.ListenAndServe(config.PORT, mux))


}
