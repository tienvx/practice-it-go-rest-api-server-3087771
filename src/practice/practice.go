package main

import (
	"example.com/backend"
)

func main() {
	backend := backend.Backend{Addr: ":9003"}
	backend.Run()
}
