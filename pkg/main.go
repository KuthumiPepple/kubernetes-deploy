package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

const (
	API_PATH = "/apis/v1/books"
)

type Book struct {
	Id, Name, Isbn string
}

type library struct {
	dbHost, dbPass, dbName string
}

func main() {
	// DB_HOST is of form host:port
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost:3306"
	}

	dbPass := os.Getenv("DB_PASS")
	if dbPass == "" {
		dbPass = "kuthumipepple"
	}

	apiPath := os.Getenv("API_PATH")
	if apiPath == "" {
		apiPath = API_PATH
	}

	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "library"
	}

	l := library{
		dbHost: dbHost,
		dbPass: dbPass,
		dbName: dbName,
	}

	r := mux.NewRouter()
	r.HandleFunc(apiPath, l.getBooks).Methods(http.MethodGet)
	r.HandleFunc(apiPath, l.postBooks).Methods(http.MethodPost)
	http.ListenAndServe(":8080", r)
}

func (l library) postBooks(w http.ResponseWriter, r *http.Request) {
	book := Book{}
	json.NewDecoder(r.Body).Decode(&book)

	db := l.openConnection()
	insertQuery, err := db.Prepare("insert into books values(?, ?, ?)")
	if err != nil {
		log.Fatalf("while preparing the db query %s\n", err.Error())
	}

	tx, err := db.Begin()
	if err != nil {
		log.Fatalf("while beginning the transaction %s\n", err.Error())
	}

	_, err = tx.Stmt(insertQuery).Exec(book.Id, book.Name, book.Isbn)
	if err != nil {
		log.Fatalf("while execing the insert command %s\n", err.Error())
	}

	err = tx.Commit()
	if err != nil {
		log.Fatalf("while committing the transaction %s\n", err.Error())
	}

	l.closeConnection(db)
}
func (l library) getBooks(w http.ResponseWriter, r *http.Request) {
	db := l.openConnection()
	rows, err := db.Query("select * from books")
	if err != nil {
		log.Fatalf("querying the books table %s\n", err.Error())
	}

	books := []Book{}
	for rows.Next() {
		var id, name, isbn string
		err := rows.Scan(&id, &name, &isbn)
		if err != nil {
			log.Fatalf("while scanning the row %s\n", err.Error())
		}
		aBook := Book{
			Id:   id,
			Name: name,
			Isbn: isbn,
		}
		books = append(books, aBook)
	}
	json.NewEncoder(w).Encode(books)
	l.closeConnection(db)
}

func (l library) openConnection() *sql.DB {
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@(%s)/%s", "root", l.dbPass, l.dbHost, l.dbName))
	if err != nil {
		log.Fatalf("opening the connection to the database %s\n", err.Error())
	}
	return db
}

func (l library) closeConnection(db *sql.DB) {
	if err := db.Close(); err != nil {
		log.Fatalf("closing connection %s\n", err.Error())
	}

}
