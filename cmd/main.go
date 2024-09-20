package main

import (
	"github.com/BazaarTrade/OrderMatchingService/internal/app"
)

func main() {
	app.Run()
}

// TOFIX:
// order shouldn`t save to the db if not enough volume
