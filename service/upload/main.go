package main

import (
	"Distributed-cloud-storage-system/route"
)

func main() {
	router := route.Router()
	router.Run(":8080")
}
