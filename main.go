package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/mgo.v2/bson"
)

type Book struct {
	ID       bson.ObjectId `json:"id" bson:"_id,omitempty"`
	Isbn     string        `json:"isbn" bson:"isbn,omitempty"`
	Title    string        `json:"title" bson:"title,omitempty"`
	Author   *Author       `json:"author" bson:"author,omitempty"`
	Location []*Location   `json:"location" bson:"location,omitempty"`
}

type Author struct {
	Firstname string `json:"firstname" bson:"firstname,omitempty"`
	Lastname  string `json:"lastname" bson:"lastname,omitempty"`
}

type Location struct {
	Code string `json:"code" bson:"code,omitempty"`
	Name string `json:"name" bson:"name,omitempty"`
}

var client *mongo.Client

var books []Book

func getBooks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(books)
}

func getBook(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	var returnBooks []Book

	collection := client.Database("NewBook").Collection("Book")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	cursor, err := collection.Find(ctx, bson.M{"isbn": params["isbn"]})

	defer cursor.Close(ctx)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message": "` + err.Error() + `"}`))
		return
	}

	for cursor.Next(ctx) {
		var book Book
		cursor.Decode(&book)
		returnBooks = append(returnBooks, book)
	}

	if err := cursor.Err(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message": "` + err.Error() + `"}`))
		return
	}
	json.NewEncoder(w).Encode(returnBooks)
	//json.NewEncoder(w).Encode(&Book{})
}

func createBook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var book Book
	_ = json.NewDecoder(r.Body).Decode(&book)

	fmt.Println(r.Body)
	collection := client.Database("NewBook").Collection("Book")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	result, _ := collection.InsertOne(ctx, book)

	//books = append(books, book)
	json.NewEncoder(w).Encode(result)
}

func updateBook(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)

	collection := client.Database("NewBook").Collection("Book")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	var book Book
	_ = json.NewDecoder(r.Body).Decode(&book)

	initQuery := bson.M{"isbn": params["isbn"]}

	updateQuery := bson.M{
		"$set": bson.M{
			"title":            book.Title,
			"author.firstname": book.Author.Firstname,
			"author.lastname":  book.Author.Lastname,
		},
	}

	fmt.Println(initQuery)

	fmt.Println(updateQuery)

	cursor, err := collection.UpdateMany(ctx, initQuery, updateQuery)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message": "` + err.Error() + `"}`))
		return
	}

	json.NewEncoder(w).Encode(cursor.ModifiedCount)
}

func updateBookAndLocationCode(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)

	collection := client.Database("NewBook").Collection("Book")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	var book Book
	_ = json.NewDecoder(r.Body).Decode(&book)

	initQuery := bson.M{"isbn": params["isbn"], "location.code": params["code"]}

	updateQuery := bson.M{
		"$set": bson.M{
			"title":            book.Title,
			"author.firstname": book.Author.Firstname,
			"author.lastname":  book.Author.Lastname,
			"location.$.name":  params["newLoc"],
		},
	}

	fmt.Println(initQuery)

	fmt.Println(updateQuery)

	cursor, err := collection.UpdateMany(ctx, initQuery, updateQuery)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message": "` + err.Error() + `"}`))
		return
	}

	json.NewEncoder(w).Encode(cursor.ModifiedCount)
}

func deleteBook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)

	for index, item := range books {
		if item.ID.String() == params["id"] {
			fmt.Println("hello world")
			books = append(books[:index], books[index+1:]...)
			break
		}
	}
	json.NewEncoder(w).Encode(books)
}

func main() {
	fmt.Println("hello world")

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	clientOptions := options.Client().ApplyURI("mongodb://mongo-service:27017")
	client, _ = mongo.Connect(ctx, clientOptions)

	r := mux.NewRouter()

	books = append(books, Book{ID: "1", Isbn: "23424", Title: "Book One", Author: &Author{Firstname: "John", Lastname: "Doe"}})
	books = append(books, Book{ID: "2", Isbn: "5435543", Title: "Book Two", Author: &Author{Firstname: "Steve", Lastname: "Smith"}})

	r.HandleFunc("/api/books", getBooks).Methods("GET")
	r.HandleFunc("/api/books/{isbn}", getBook).Methods("GET")
	r.HandleFunc("/api/books", createBook).Methods("POST")
	r.HandleFunc("/api/books/{isbn}", updateBook).Methods("PUT")
	r.HandleFunc("/api/books/{isbn}/{code}/{newLoc}", updateBookAndLocationCode).Methods("PUT")
	r.HandleFunc("/api/books/{id}", deleteBook).Methods("DELETE")

	log.Fatal(http.ListenAndServe(":8000", r))
}
