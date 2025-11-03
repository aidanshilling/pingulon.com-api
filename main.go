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

func obtainMongoClient() *mongo.Client {
	godotenv.Load()
	uri := os.Getenv("MONGO_URI")

	if uri == "" {
		log.Fatal("Set your 'MONGO_URI' environment variable.")
	}

	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		// TODO: Remove these panic's
		panic(err)
	}
	return client
}

func getArticle(client *mongo.Client, name string, collection string) json.RawMessage {
	coll := client.Database("articles").Collection(collection)

	var result bson.M
	err := coll.FindOne(context.TODO(), bson.D{{"name", name}}).Decode(&result)
	if err == mongo.ErrNoDocuments {
		fmt.Printf("No document was found with the name %s\n", name)
		return nil
	}
	if err != nil {
		panic(err)
	}

	jsonData, err := json.MarshalIndent(result, "", "	")
	if err != nil {
		panic(err)
	}

	return jsonData
}

func main() {
	mongoClient := obtainMongoClient()
	defer func() {
		if err := mongoClient.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()

	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		return c.JSON(http.StatusOK, getArticle(mongoClient, "Testy", "stories"))
	})
	e.Logger.Fatal(e.Start(":1323"))
}
