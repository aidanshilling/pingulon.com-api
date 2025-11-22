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
	"github.com/labstack/echo/v4/middleware"
)

type Article struct {
	ID   bson.ObjectID `bson:"_id" json:"id"`
	Name string        `bson:"name" json:"name"`
}

func obtainMongoClient() (*mongo.Client, error) {
	godotenv.Load()
	uri := os.Getenv("MONGO_URI")

	if uri == "" {
		log.Fatal("Set your 'MONGO_URI' environment variable.")
	}

	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}
	return client, nil
}

func getArticleByName(client *mongo.Client, name string, collection string) (json.RawMessage, error) {
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

func getArticleById(client *mongo.Client, idStr string, collection string) (json.RawMessage, error) {
	coll := client.Database("articles").Collection(collection)

	id, err := bson.ObjectIDFromHex(idStr)
	if err != nil {
		return nil, fmt.Errorf("invalid object id: %w", err)
	}

	var result bson.M
	err = coll.FindOne(context.TODO(), bson.D{{"_id", id}}).Decode(&result)
	if err == mongo.ErrNoDocuments {
		fmt.Printf("No document was found with the id %s\n", id)
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

func getArticleList(client *mongo.Client, collection string) ([]Article, error) {
	ctx := context.TODO()
	coll := client.Database("articles").Collection(collection)
	projection := bson.D{{Key: "_id", Value: 1}, {Key: "name", Value: 1}}
	opts := options.Find().SetProjection(projection)

	cursor, err := coll.Find(ctx, bson.D{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var articles []Article
	if err := cursor.All(ctx, &articles); err != nil {
		return nil, err
	}

	return articles, nil
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
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"http://localhost:5173"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Please search for an article with format /:collection/:name")
	})
	e.GET("/:collection", func(c echo.Context) error {
		collection := c.Param("collection")
		articles, err := getArticleList(mongoClient, collection)
		if err != nil {
			return c.String(http.StatusBadRequest, "Failed to fetch collection\n")
		}
		return c.JSON(http.StatusOK, articles)

	})
	e.GET("/:collection/name/:name", func(c echo.Context) error {
		collection := c.Param("collection")
		name := c.Param("name")
		articleJson, err := getArticleByName(mongoClient, name, collection)
		if err != nil {
			return c.String(http.StatusBadRequest, "Failed to fetch article\n")
		}
		return c.JSON(http.StatusOK, articleJson)
	})
	e.GET("/:collection/id/:id", func(c echo.Context) error {
		collection := c.Param("collection")
		id := c.Param("id")
		articleJson, err := getArticleById(mongoClient, id, collection)
		if err != nil {
			return c.String(http.StatusBadRequest, "Failed to fetch article\n")
		}
		return c.JSON(http.StatusOK, articleJson)
	})
	e.Logger.Fatal(e.Start(":8080"))
}
