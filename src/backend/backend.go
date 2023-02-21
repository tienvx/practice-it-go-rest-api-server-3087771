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
	DB     *sql.DB
	Addr   string
	Router *mux.Router
}

func (b *Backend) Init() error {
	db, err := sql.Open("sqlite3", "../../practiceit.db")
	if err != nil {
		return fmt.Errorf("Can not connect database: %s", err)
	}
	b.DB = db
	b.initRoutes()
	return nil
}

func (b Backend) Run() {
	b.Init()
	http.Handle("/", b.Router)
	fmt.Println("Server started and listening on port ", b.Addr)
	log.Fatal(http.ListenAndServe(b.Addr, nil))
}

func (b *Backend) initRoutes() {
	router := mux.NewRouter()
	router.HandleFunc("/products", b.allProducts).Methods("GET")
	router.HandleFunc("/products/{id}", b.fetchProduct).Methods("GET")
	router.HandleFunc("/products", b.newProduct).Methods("POST")
	router.HandleFunc("/orders", b.allOrders).Methods("GET")
	router.HandleFunc("/orders/{id}", b.fetchOrder).Methods("GET")
	router.HandleFunc("/orders", b.newOrder).Methods("POST")

	b.Router = router
}

func (b *Backend) allProducts(w http.ResponseWriter, r *http.Request) {
	products, err := getProducts(b.DB)
	if err != nil {
		fmt.Printf("Can not get all products: %s\n", err.Error())
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, products)
}

func (b *Backend) fetchProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var p product
	p.Id, _ = strconv.Atoi(id)
	err := p.getProduct(b.DB)
	if err != nil {
		fmt.Printf("Can not get product: %s\n", err.Error())
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, p)
}

func (b *Backend) newProduct(w http.ResponseWriter, r *http.Request) {
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("Can not create product: %s\n", err.Error())
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	var p product
	err = json.Unmarshal(reqBody, &p)
	if err != nil {
		fmt.Printf("Can not create product: %s\n", err.Error())
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	err = p.createProduct(b.DB)
	if err != nil {
		fmt.Printf("Can not create product: %s\n", err.Error())
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, p)
}

func (b *Backend) allOrders(w http.ResponseWriter, r *http.Request) {
	orders, err := getOrders(b.DB)
	if err != nil {
		fmt.Printf("Can not get all orders: %s\n", err.Error())
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, orders)
}

func (b *Backend) fetchOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var o order
	o.Id, _ = strconv.Atoi(id)
	err := o.getOrder(b.DB)
	if err != nil {
		fmt.Printf("Can not get order: %s\n", err.Error())
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, o)
}

func (b *Backend) newOrder(w http.ResponseWriter, r *http.Request) {
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("Can not create order: %s\n", err.Error())
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	var o order
	err = json.Unmarshal(reqBody, &o)
	if err != nil {
		fmt.Printf("Can not create order: %s\n", err.Error())
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	err = o.createOrder(b.DB)
	if err != nil {
		fmt.Printf("Can not create order: %s\n", err.Error())
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, o)
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
