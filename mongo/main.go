package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type User struct {
	Name  string `bson:"name"`
	Email string `bson:"email"`
	Age   int    `bson:"age"`
}

func main() {
	// Подключение к MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	// Проверка подключения
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connected to MongoDB!")

	// Выбор базы данных и коллекции
	db := client.Database("testdb")
	usersCollection := db.Collection("users")

	// Вставка документа
	user := User{Name: "Иван", Email: "ivan@example.com", Age: 30}
	insertResult, err := usersCollection.InsertOne(ctx, user)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Inserted ID:", insertResult.InsertedID)

	// Поиск документа
	var result User
	err = usersCollection.FindOne(ctx, bson.M{"name": "Иван"}).Decode(&result)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Found user: %+v\n", result)
}
