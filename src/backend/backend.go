package backend

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	_ "github.com/mattn/go-sqlite3"
)

type Product struct {
	id        int
	name      string
	inventory int
	price     int
}

type Backend struct {
	db   *sql.DB
	Addr string
}

func (b *Backend) Open(file string) error {
	db, err := sql.Open("sqlite3", file)
	if err != nil {
		return fmt.Errorf("Can not connect database: %s", err)
	}
	b.db = db
	return nil
}

func (b Backend) Fetch() ([]Product, error) {
	rows, err := b.db.Query("SELECT id, name, inventory, price FROM products")
	if err != nil {
		return nil, fmt.Errorf("Can not query: %s", err)
	}

	defer rows.Close()

	var products []Product
	for rows.Next() {
		var p Product

		rows.Scan(&p.id, &p.name, &p.inventory, &p.price)

		products = append(products, p)
		fmt.Printf("Product: %d, %s, %d, %d\n", p.id, p.name, p.inventory, p.price)
	}

	return products, nil
}

func (b Backend) Run() {
	r := mux.NewRouter()
	r.HandleFunc("/products", b.getProducts).Methods("GET")
	http.Handle("/", r)
	fmt.Println("Server started and listening on port ", b.Addr)
	log.Fatal(http.ListenAndServe(b.Addr, nil))
}

func (b Backend) getProducts(rw http.ResponseWriter, r *http.Request) {
	err := b.Open("../../practiceit.db")
	if err != nil {
		fmt.Fprintf(rw, "Can not open database: %s", err)
		return
	}
	products, err := b.Fetch()
	if err != nil {
		fmt.Fprintf(rw, "Can not fetch products: %s", err)
		return
	}
	fmt.Fprintf(rw, "Products:\n%v", products)
}
