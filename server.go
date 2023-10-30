package main

import (
	"database/sql"
	"fmt"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"
	"graphql_book_management/graph"
	"log"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/joho/godotenv"
	"github.com/uptrace/bun"
)

const defaultPort = "4000"

var DB *bun.DB

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	err := connectToDatabase()

	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Connected to database\n")

	resolver := &graph.Resolver{DB: DB}

	executableSchema := graph.NewExecutableSchema(graph.Config{Resolvers: resolver})
	server := handler.NewDefaultServer(executableSchema)

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", server)

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func connectToDatabase() error {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Build the database connection string
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_SSL_MODE"),
	)

	// Open a connection to the PostgreSQL database
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))

	// Initialize the bun.DB instance with the PostgreSQL dialect
	DB = bun.NewDB(sqldb, pgdialect.New())

	// Add a query hook for debugging purposes
	DB.AddQueryHook(bundebug.NewQueryHook(
		bundebug.WithVerbose(true),
		bundebug.FromEnv("BUNDEBUG"),
	))

	return DB.Ping()
}
