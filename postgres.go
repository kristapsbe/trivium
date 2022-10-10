package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"
)

func connectToDb() {
	viper.AddConfigPath("./")
	viper.SetConfigName("config") // Register config file name (no extension)
	viper.SetConfigType("json")   // Look for specific type
	CheckError(viper.ReadInConfig())

	host := viper.Get("db.prod.host")
	port := viper.GetInt("db.prod.port")
	user := viper.Get("db.prod.user")
	password := viper.Get("db.prod.password")
	dbname := viper.Get("db.prod.dbname")

	// connection string
	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	// open database
	db, err := sql.Open("postgres", psqlconn)
	CheckError(err)

	rows, err := db.Query(`SELECT "id", "moves" FROM "game"`)
	CheckError(err)

	defer rows.Close()
	for rows.Next() {
		var id string
		var moves string

		err = rows.Scan(&id, &moves)
		CheckError(err)

		fmt.Println(id, moves)
	}

	CheckError(err)

	// close database
	defer db.Close()

	// check db
	err = db.Ping()
	CheckError(err)

	fmt.Println("Connected to DB")
}

func CheckError(err error) {
	if err != nil {
		panic(err)
	}
}
