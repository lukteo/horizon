package main

import (
	"github.com/luketeo/horizon/internal/boot"
	"github.com/luketeo/horizon/internal/config"
)

func main() {
	c := config.NewConfig()
	s := boot.NewServer(c)

	s.Start()
}
