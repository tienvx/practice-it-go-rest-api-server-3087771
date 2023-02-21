package backend

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	_ "github.com/mattn/go-sqlite3"
)

type Backend struct {
	db     *sql.DB
	Addr   string
	router *mux.Router
}

func (b *Backend) init() error {
	db, err := sql.Open("sqlite3", "../../practiceit.db")
	if err != nil {
		return fmt.Errorf("Can not connect database: %s", err)
	}
	b.db = db
	b.initRoutes()
	return nil
}

func (b Backend) Run() {
	b.init()
	http.Handle("/", b.router)
	fmt.Println("Server started and listening on port ", b.Addr)
	log.Fatal(http.ListenAndServe(b.Addr, nil))
}

func (b *Backend) initRoutes() {
	router := mux.NewRouter()
	router.HandleFunc("/products", b.allProducts).Methods("GET")
	router.HandleFunc("/products/{id}", b.fetchProduct).Methods("GET")
	router.HandleFunc("/products", b.newProduct).Methods("POST")

	b.router = router
}

func (b *Backend) allProducts(rw http.ResponseWriter, r *http.Request) {
	products, err := getProducts(b.db)
	if err != nil {
		fmt.Printf("Can not get all products: %s", err.Error())
		respondWithError(rw, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(rw, http.StatusOK, products)
}

func (b *Backend) fetchProduct(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var p product
	p.Id, _ = strconv.Atoi(id)
	err := p.getProduct(b.db)
	if err != nil {
		fmt.Printf("Can not get product: %s", err.Error())
		respondWithError(rw, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(rw, http.StatusOK, p)
}

func (b *Backend) newProduct(rw http.ResponseWriter, r *http.Request) {
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("Can not create product: %s", err.Error())
		respondWithError(rw, http.StatusBadRequest, err.Error())
		return
	}
	var p product
	err = json.Unmarshal(reqBody, &p)
	if err != nil {
		fmt.Printf("Can not create product: %s", err.Error())
		respondWithError(rw, http.StatusInternalServerError, err.Error())
		return
	}
	err = p.createProduct(b.db)
	if err != nil {
		fmt.Printf("Can not create product: %s", err.Error())
		respondWithError(rw, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(rw, http.StatusOK, p)
}

// Helper functions
func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
