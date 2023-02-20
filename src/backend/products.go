package backend

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

type product struct {
	Id          int    `json:"id"`
	ProductCode string `json:"productCode"`
	Name        string `json:"name"`
	Inventory   int    `json:"inventory"`
	Price       int    `json:"price"`
	Status      string `json:"status"`
}

func getProducts(db *sql.DB) ([]product, error) {
	rows, err := db.Query("SELECT id, productCode, name, inventory, price, status FROM products")
	if err != nil {
		return nil, fmt.Errorf("Can not query: %s", err)
	}

	defer rows.Close()

	var products []product
	for rows.Next() {
		var p product

		err = rows.Scan(&p.Id, &p.ProductCode, &p.Name, &p.Inventory, &p.Price, &p.Status)
		if err != nil {
			return nil, fmt.Errorf("Can not scan row: %s", err)
		}

		products = append(products, p)
	}

	return products, nil
}

func (p *product) getProduct(db *sql.DB) error {
	return db.QueryRow("SELECT productCode, name, inventory, price, status FROM products WHERE id = ?", p.Id).Scan(&p.ProductCode, &p.Name, &p.Inventory, &p.Price, &p.Status)
}
