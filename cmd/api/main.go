package main

import (
	"app/internal/data"
	"app/internal/woodlog"
	"context"
	"flag"
	"fmt"
	"html/template"
	"os"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type application struct {
	config        config
	models        data.Models
	logger        *woodlog.Logger
	templateCache map[string]*template.Template

	wg sync.WaitGroup
}

type config struct {
	port string
	db   struct {
		dns string
	}
}

func main() {
	godotenv.Load()
	var config config
	flag.StringVar(&config.port, "port", os.Getenv("PORT"), "port")
	fmt.Println(os.Getenv("PORT"))
	// flag.StringVar(&config.db.dns, "uri", os.Getenv("MONGOURI"), "mongo uri")
	flag.StringVar(&config.db.dns, "uri", "mongodb://localhost:27017/advanced", "mongo uri")

	logger := *woodlog.New(os.Stdout, 0)

	templateCache, err := data.NewTemplateCache("./ui/html/")
	if err != nil {
		logger.PrintFatal(err.Error(), "failed to create template cache")
	}

	db := mustOpenDB(config.db.dns)
	defer db.Client().Disconnect(context.TODO())

	app := application{
		templateCache: templateCache,
		config:        config,
		logger:        &logger,
		models:        data.NewModels(db),
	}

	err = app.serve()
	if err != nil {
		logger.PrintFatal(err.Error(), "failed to start server")
	}
}

// fix later: hardcoded collection name
func mustOpenDB(dsn string) *mongo.Database {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	db, err := mongo.Connect(ctx, options.Client().ApplyURI(dsn))
	if err != nil {
		panic(err)
	}
	_, err = db.Database("novye").Collection("users").Indexes().CreateMany(
		context.Background(),
		[]mongo.IndexModel{
			{
				Keys:    bson.D{{Key: "login", Value: 1}},
				Options: options.Index().SetUnique(true),
			},
			{
				Keys:    bson.D{{Key: "email", Value: 1}},
				Options: options.Index().SetUnique(true),
			},
		},
	)
	if err != nil {
		panic(err)
	}
	return db.Database("novye")
}
