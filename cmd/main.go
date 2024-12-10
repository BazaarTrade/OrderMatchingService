package main

import (
	"time"

	"github.com/BazaarTrade/OrderMatchingService/internal/app"
)

func main() {
	time.Sleep(time.Second * 3)
	app.Run()
}
