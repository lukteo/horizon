package main

import (
	"github.com/luketeo/horizon/config"
	"github.com/luketeo/horizon/internal/boot"
)

func main() {
	c := config.NewConfig()
	b := boot.NewAPI(c)

	b.Start()
}
