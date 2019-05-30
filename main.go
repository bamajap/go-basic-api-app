/*
Author: Jason Payne
*/
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	// Run the app in "test" mode.
	db "go-basic-api-app/dummydb"

	// Run the app with DynamoDB.
	// db "go-basic-api-app/dynamodb"
	"strconv"

	"github.com/gorilla/mux"
)

/*
GetAllProducts - display all of the Products.
*/
func GetAllProducts(w http.ResponseWriter, r *http.Request) {
	p, err := db.Items.GetAll()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(p)
}

/*
CreateProduct - create a new Product and add to the database.
*/
func CreateProduct(w http.ResponseWriter, r *http.Request) {
	var p db.Product

	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	defer r.Body.Close()

	if err := db.Items.AddProduct(p); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(p)
}

/*
GetProduct - display a single Product based on ID or Name.
*/
func GetProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	p := db.Product{Id: id}
	if err = db.Items.GetProduct(&p); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(p)
}

/*
UpdateProduct - update an existing Product.
*/
func UpdateProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var p db.Product

	if err = json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	defer r.Body.Close()

	p.Id = id

	if err = db.Items.UpdateProduct(p); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(p)
}

/*
DeleteProduct - delete a Product from the database.
*/
func DeleteProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	p := db.Product{Id: id}
	if err = db.Items.DeleteProduct(p); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"result": "success"})
}

func main() {
	fmt.Println("Initializing database...")
	if initErr := db.Initialize(); initErr != nil {
		if cleanupErr := db.Cleanup(); cleanupErr != nil {
			fmt.Println(cleanupErr.Error())
		}
		log.Fatal(initErr.Error())
	}

	// Ideally, the server shutdown would get handled gracefully allowing for post-shutdown cleanup tasks like below.
	// defer func() {
	// 	if cleanupErr := db.Cleanup(); cleanupErr != nil {
	// 		fmt.Println(cleanupErr.Error())
	// 	}
	// }()

	fmt.Println("DONE!")

	router := mux.NewRouter()
	router.HandleFunc("/", GetAllProducts).Methods(http.MethodGet)
	router.HandleFunc("/product", CreateProduct).Methods(http.MethodPost)
	router.HandleFunc("/product/{id:[0-9]+}", GetProduct).Methods(http.MethodGet)
	router.HandleFunc("/product/{id:[0-9]+}", UpdateProduct).Methods(http.MethodPut)
	router.HandleFunc("/product/{id:[0-9]+}", DeleteProduct).Methods(http.MethodDelete)

	// http://localhost:8000
	log.Fatal(http.ListenAndServe(":8000", router))
}
