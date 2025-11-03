package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"net/http"

	"github.com/labstack/echo/v4"
)

func obtainMongoClient() (*mongo.Client, error) {
	godotenv.Load()
	uri := os.Getenv("MONGO_URI")

	if uri == "" {
		log.Fatal("Set your 'MONGO_URI' environment variable.")
	}

	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		// TODO: Remove these panic's
		return nil, err
	}
	return client, nil
}

func getArticle(client *mongo.Client, name string, collection string) (json.RawMessage, error) {
	coll := client.Database("articles").Collection(collection)

	var result bson.M
	err := coll.FindOne(context.TODO(), bson.D{{"name", name}}).Decode(&result)
	if err == mongo.ErrNoDocuments {
		fmt.Printf("No document was found with the name %s\n", name)
		return nil, err
	}
	if err != nil {
		return nil, err
	}

	jsonData, err := json.MarshalIndent(result, "", "	")
	if err != nil {
		return nil, err
	}

	return jsonData, nil
}

func main() {
	mongoClient, err := obtainMongoClient()
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := mongoClient.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()

	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		articleJson, err := getArticle(mongoClient, "Bunion", "other")
		if err != nil {
			return c.String(http.StatusBadRequest, "Failed to fetch article")
		}
		return c.JSON(http.StatusOK, articleJson)
	})
	e.Logger.Fatal(e.Start(":1323"))
}
