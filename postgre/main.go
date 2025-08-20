package main

import (
	"database/sql"
	"fmt"
	"log"
	"sync"

	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "postgres"
	dbname   = "postgres"
)

type (
	User struct {
		gorm.Model
		ID   uint   `gorm:"primaryKey"`
		Name string `gorm:"type:varchar(50)"`
		Age  int    `gorm:"default:18"`
	}
)

func GORMmain() {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := gorm.Open(postgres.Open(psqlInfo), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	db.AutoMigrate(&User{})

	user := User{Name: "Tim Mallen", Age: 17}
	result := db.Create(&user)

	if result.Error != nil {
		log.Fatal(err)
	}
}

func main() {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Успешное подключение к PostgreSQL!")

	_, err = db.Exec("DROP TABLE test")
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`
        CREATE TABLE test (
            id INT PRIMARY KEY,
            name TEXT,
            email TEXT
        )
    `)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec("INSERT INTO test (id, name, email) VALUES ($1, $2, $3)", 1, "Иван", "ivan@test.com")
	if err != nil {
		log.Fatal(err)
	}

	rows, err := db.Query("SELECT id, name, email FROM test")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var name, email string
		err = rows.Scan(&id, &name, &email)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("ID: %d, Имя: %s, Email: %s\n", id, name, email)
	}

	_, err = db.Exec("DROP TABLE account")
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS account (
            id INT PRIMARY KEY,
            balance INT
        )
    `)

	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec("INSERT INTO account (id, balance) VALUES($1, $2)", 1, 0)

	if err != nil {
		log.Fatal(err)
	}

	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, err = tx.Exec("UPDATE account SET balance = 100 WHERE id = $1", 1)
		if err != nil {
			tx.Rollback()
			log.Fatal(err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		_, err = tx.Exec("UPDATE account SET balance = 101 WHERE id = $1", 1)
		if err != nil {
			tx.Rollback()
			log.Fatal(err)
		}
	}()

	wg.Wait()

	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	}

	rows2, err := db.Query("SELECT * FROM account")
	if err != nil {
		log.Fatal(err)
	}

	defer rows2.Close()
	for rows2.Next() {
		var id, balance int
		err = rows2.Scan(&id, &balance)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("ID = %d, balance = %d\n", id, balance)
	}
}
