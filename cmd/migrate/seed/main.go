package main

import (
	"fmt"
	"log"

	"github.com/Shadowcyng/goSocial/internal/db"
	"github.com/Shadowcyng/goSocial/internal/store"
)

func main() {
	conn, err := db.New("postgres://satyam:satyam123@localhost/social?sslmode=require", 3, 3, "15m")
	if err != nil {
		log.Fatal("Connection failed", err)
	}

	if err != nil {
		log.Panic(err)
	}
	defer conn.Close()
	fmt.Println("database connection pool established")
	store := store.NewStorage(conn)
	db.Seed(store)

}
