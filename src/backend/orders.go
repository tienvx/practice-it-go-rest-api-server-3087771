package backend

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

type order struct {
	Id           int         `json:"id"`
	CustomerName string      `json:"customerName"`
	Total        int         `json:"total"`
	Status       string      `json:"status"`
	Items        []orderItem `json:"items"`
}

type orderItem struct {
	ProductId int `json:"productId"`
	Quantity  int `json:"quantity"`
}

func getOrders(db *sql.DB) ([]order, error) {
	rows, err := db.Query("SELECT id, customerName, total, status FROM orders")
	if err != nil {
		return nil, fmt.Errorf("Can not query: %s", err)
	}

	defer rows.Close()

	var orders []order
	for rows.Next() {
		var o order

		err = rows.Scan(&o.Id, &o.CustomerName, &o.Total, &o.Status)
		if err != nil {
			return nil, fmt.Errorf("Can not scan row: %s", err)
		}
		err = o.getItems(db)
		if err != nil {
			return nil, fmt.Errorf("Can not get items for order: %s", &err)
		}

		orders = append(orders, o)
	}

	return orders, nil
}

func (o *order) getOrder(db *sql.DB) error {
	err := db.QueryRow("SELECT customerName, total, status FROM orders WHERE id = ?", o.Id).Scan(&o.CustomerName, &o.Total, &o.Status)
	if err != nil {
		return err
	}
	err = o.getItems(db)
	if err != nil {
		return err
	}
	return nil
}

func (o *order) createOrder(db *sql.DB) error {
	res, err := db.Exec("INSERT INTO orders(customerName, total, status) VALUES (?, ?, ?)", o.CustomerName, o.Total, o.Status)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	o.Id = int(id)
	err = o.createItems(db)
	if err != nil {
		return err
	}
	return nil
}

func (o *order) createItems(db *sql.DB) error {
	for _, item := range o.Items {
		_, err := db.Exec("INSERT INTO order_items(order_id, product_id, quantity) VALUES (?, ?, ?)", o.Id, item.ProductId, item.Quantity)
		if err != nil {
			return err
		}
	}
	return nil
}

func (o *order) getItems(db *sql.DB) error {
	rows, err := db.Query("SELECT product_id, quantity FROM order_items WHERE order_id = ?", o.Id)
	if err != nil {
		return fmt.Errorf("Can not query: %s", err)
	}

	defer rows.Close()

	var items []orderItem
	for rows.Next() {
		var i orderItem

		err = rows.Scan(&i.ProductId, &i.Quantity)
		if err != nil {
			return fmt.Errorf("Can not scan row: %s", err)
		}

		items = append(items, i)
	}

	o.Items = items
	return nil
}
