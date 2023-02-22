package main

import (
    "fmt"
    "log"
    "net/http"
	 "go.mongodb.org/mongo-driver/bson"
)



func make_bson() bson.D{ 
	return bson.D{{"user_id","11"},{"encrypted_key","11"}}

}


func formHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		 fmt.Fprintf(w, "ParseForm() err: %v", err)
		 return
	}
	fmt.Fprintf(w, "POST request successful")
	name := r.FormValue("name")
	address := r.FormValue("address")

	fmt.Fprintf(w, "Name = %s\n", name)
	fmt.Fprintf(w, "Address = %s\n", address)
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/hello" {
		 http.Error(w, "404 not found.", http.StatusNotFound)
		 return
	}

	if r.Method != "GET" {
		 http.Error(w, "Method is not supported.", http.StatusNotFound)
		 return
	}


	fmt.Fprintf(w, "Hello!")
}

func main() {
	//read("TEST", "10"
//hget request for john)
//if read("TEST", "JOHN")

	//return
	fileServer := http.FileServer(http.Dir("./static")) // New code
	http.Handle("/", fileServer) // New code
	http.HandleFunc("/hello", helloHandler)
	http.HandleFunc("/form", formHandler)


	fmt.Printf("Starting server at port 8080\n")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		 log.Fatal(err)
	}
}