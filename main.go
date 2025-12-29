/*
Copyright © 2025 Marouane Boufarouj <boufaroujmarouan@gmail.com>
*/
package main

import (
	"fmt"
	"log"

	"github.com/chibuka/95-cli/internal/config"
)

func main() {
	config.Init()

	// cmd.Execute()
	cfg := config.Config{
		APIUrl:       "api",
		AccessToken:  "token",
		RefreshToken: "refresh_token",
		UserId:       3,
		Username:     "grainme",
	}

	err := cfg.Save()
	if err != nil {
		log.Fatal(err)
	}

	cfgLoaded, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Config loaded: ", *cfgLoaded)

	err = config.Clear()
	if err != nil {
		log.Fatal(err)
	}

}
