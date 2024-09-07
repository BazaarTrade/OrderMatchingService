package main

import (
	"github.com/Moha192/OrderMatchingService/internal/app"
)

func main() {
	app.Run()
}

// TOFIX:
// order shouldn`t save to the db if not enough volume
