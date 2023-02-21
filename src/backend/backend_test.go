package backend_test

import (
	"os"
	"testing"

	"bytes"
	"encoding/json"
	"example.com/backend"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
)

var b backend.Backend

const tableProductCreationQuery = `CREATE TABLE IF NOT EXISTS products
(
	id INT NOT NULL PRIMARY KEY AUTOINCREMENT,
	productCode VARCHAR(25) NOT NULL,
	name VARCHAR(256) NOT NULL,
	inventory INT NOT NULL,
	price INT NOT NULL,
	status VARCHAR(64) NOT NULL
)`

const tableOrderCreationQuery = `CREATE TABLE IF NOT EXISTS orders
(
	id INT NOT NULL PRIMARY KEY AUTOINCREMENT,
	customerName VARCHAR(256) NOT NULL,
	total INT NOT NULL,
	status VARCHAR(64) NOT NULL
)`

const tableOrderItemsCreationQuery = `CREATE TABLE IF NOT EXISTS order_items
(
	order_id INT,
	product_id INT,
	quantity INT NOT NULL,
	FOREIGN KEY (order_id) REFERENCES orders (id),
	FOREIGN KEY (product_id) REFERENCES products (id),
	PRIMARY KEY (order_id, product_id)
)`

func TestMain(m *testing.M) {
	b = backend.Backend{}
	b.Init()
	ensureTableExists()
	code := m.Run()

	clearProductTable()
	clearOrderTable()
	clearOrderItemsTable()
	os.Exit(code)
}

func ensureTableExists() {
	if _, err := b.DB.Exec(tableProductCreationQuery); err != nil {
		log.Fatal(err)
	}
	if _, err := b.DB.Exec(tableOrderCreationQuery); err != nil {
		log.Fatal(err)
	}
	if _, err := b.DB.Exec(tableOrderItemsCreationQuery); err != nil {
		log.Fatal(err)
	}
}

func clearProductTable() {
	b.DB.Exec("DELETE FROM products")
	b.DB.Exec("DELETE FROM sqlite_sequence WHERE name = 'products'")
}

func addProduct() int {
	res, _ := b.DB.Exec("INSERT INTO products(productCode, name, inventory, price, status) VALUES (?, ?, ?, ?, ?)", "Code 123", "Name 234", 1, 2, "test")
	id, _ := res.LastInsertId()
	return int(id)
}

func TestGetNonExistentProduct(t *testing.T) {
	clearProductTable()

	req, _ := http.NewRequest("GET", "/products/11", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusInternalServerError, response.Code)

	var m map[string]string
	json.Unmarshal(response.Body.Bytes(), &m)
	if m["error"] != "sql: no rows in result set" {
		t.Errorf("Expected the 'error' key of the response to be set to 'sql: no rows in result set'. Got '%s'", m["error"])
	}
}

func TestGetProduct(t *testing.T) {
	clearProductTable()
	id := addProduct()

	req, _ := http.NewRequest("GET", fmt.Sprintf("/products/%d", id), nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	verifyProduct(t, m, 1.0, "Code 123", "Name 234", 1.0, 2.0, "test")
}

func TestCreateProduct(t *testing.T) {
	clearProductTable()

	payload := []byte(`{"productCode": "TEST12345", "name": "ProductTest", "inventory": 1, "price": 1, "status": "testing"}`)

	req, _ := http.NewRequest("POST", "/products", bytes.NewBuffer(payload))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	verifyProduct(t, m, 1.0, "TEST12345", "ProductTest", 1.0, 1.0, "testing")
}

func verifyProduct(t *testing.T, m map[string]interface{}, id, productCode, name, inventory, price, status any) {
	if m["productCode"] != productCode {
		t.Errorf("Expected productCode to be '%v'. Got '%v'", productCode, m["productCode"])
	}

	if m["name"] != name {
		t.Errorf("Expected name to be '%v'. Got '%v'", name, m["name"])
	}

	if m["inventory"] != inventory {
		t.Errorf("Expected inventory to be '%v'. Got '%v'", inventory, m["inventory"])
	}

	if m["price"] != price {
		t.Errorf("Expected price to be '%v'. Got '%v'", price, m["price"])
	}

	if m["status"] != status {
		t.Errorf("Expected status to be '%v'. Got '%v'", status, m["status"])
	}

	if m["id"] != id {
		t.Errorf("Expected id to be '%v'. Got '%v'", id, m["id"])
	}
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	b.Router.ServeHTTP(rr, req)

	return rr
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}
}

type OrderItem struct {
	productId float64
	quantity  float64
}

func clearOrderTable() {
	b.DB.Exec("DELETE FROM orders")
	b.DB.Exec("DELETE FROM sqlite_sequence WHERE name = 'orders'")
}

func clearOrderItemsTable() {
	b.DB.Exec("DELETE FROM order_items")
}

func addOrder() int {
	res, _ := b.DB.Exec("INSERT INTO orders(customerName, total, status) VALUES (?, ?, ?)", "Name 123", 3, "practice")
	id, _ := res.LastInsertId()
	b.DB.Exec("INSERT INTO order_items(order_id, product_id, quantity) VALUES (?, ?, ?)", id, 4, 6)
	b.DB.Exec("INSERT INTO order_items(order_id, product_id, quantity) VALUES (?, ?, ?)", id, 7, 12)
	return int(id)
}

func verifyOrder(t *testing.T, m map[string]interface{}, id, customerName, total, status any, items []OrderItem) {
	if m["customerName"] != customerName {
		t.Errorf("Expected customerName to be '%v'. Got '%v'", customerName, m["customerName"])
	}

	if m["total"] != total {
		t.Errorf("Expected total to be '%v'. Got '%v'", total, m["total"])
	}

	if m["status"] != status {
		t.Errorf("Expected status to be '%v'. Got '%v'", status, m["status"])
	}

	if m["id"] != id {
		t.Errorf("Expected id to be '%v'. Got '%v'", id, m["id"])
	}

	for index, item := range items {
		if getOrderItemValue(m, index, "productId") != item.productId {
			t.Errorf("Expected productId to be '%v'. Got '%v'", item.productId, getOrderItemValue(m, index, "productId"))
		}

		if getOrderItemValue(m, index, "quantity") != item.quantity {
			t.Errorf("Expected quantity to be '%v'. Got '%v'", item.quantity, getOrderItemValue(m, index, "quantity"))
		}
	}
}

func getOrderItemValue(m map[string]interface{}, index int, key string) float64 {
	return m["items"].([]interface{})[index].(map[string]interface{})[key].(float64)
}

func TestGetNonExistentOrder(t *testing.T) {
	clearOrderTable()
	clearOrderItemsTable()

	req, _ := http.NewRequest("GET", "/orders/11", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusInternalServerError, response.Code)

	var m map[string]string
	json.Unmarshal(response.Body.Bytes(), &m)
	if m["error"] != "sql: no rows in result set" {
		t.Errorf("Expected the 'error' key of the response to be set to 'sql: no rows in result set'. Got '%s'", m["error"])
	}
}

func TestGetOrder(t *testing.T) {
	clearOrderTable()
	clearOrderItemsTable()
	id := addOrder()

	req, _ := http.NewRequest("GET", fmt.Sprintf("/orders/%d", id), nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	verifyOrder(t, m, 1.0, "Name 123", 3.0, "practice", []OrderItem{{4, 6}, {7, 12}})
}

func TestCreateOrder(t *testing.T) {
	clearOrderTable()
	clearOrderItemsTable()

	payload := []byte(`{"customerName": "Customer 123", "total": 1, "status": "testing", "items": [{"productId": 1, "quantity": 2}, {"productId": 3, "quantity": 4}]}`)

	req, _ := http.NewRequest("POST", "/orders", bytes.NewBuffer(payload))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	verifyOrder(t, m, 1.0, "Customer 123", 1.0, "testing", []OrderItem{{1, 2}, {3, 4}})
}
