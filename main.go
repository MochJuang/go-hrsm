package main

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoInstance struct {
	Client *mongo.Client
	Db     *mongo.Database
}

var mg MongoInstance

const dbName = "fiber-hrms"
const mongoURI = "mongondb://localhost:27017/" + dbName

type Employee struct {
	ID     string  `json:"id, omitempty" bson:"_id, omitempty"`
	Name   string  `json:"name"`
	Salary float64 `json:"salary"`
	Age    float64 `json:"age"`
}

func Connect() error {
	client, err := mongo.NewClient(options.Client().ApplyURI(mongoURI))

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	db := client.Database(dbName)

	if err != nil {
		return err
	}

	mg = MongoInstance{
		Client: client,
		Db:     db,
	}
	return nil

}

func main() {
	app := fiber.New()

	employee := app.Group("employee")
	employee.Get("/", func(c *fiber.Ctx) error {
		query := bson.D{{}}
		cursor, err := mg.Db.Collection("employee").Find(c.Context(), query)

		if err != nil {
			return c.Status(500).SendString(err.Error())
		}

		var employees []Employee = make([]Employee, 0)

		if err := cursor.All(c.Context(), &employees); err != nil {
			return c.Status(500).SendString(err.Error())
		}

		return c.JSON(employees)
	})

	employee.Post("/", func(c *fiber.Ctx) error {
		collection := mg.Db.Collection("employee")

		var employee Employee
		if err := c.BodyParser(employee); err != nil {
			return c.Status(400).SendString(err.Error())
		}

		employee.ID = ""

		result, err := collection.InsertOne(c.Context(), employee)

		if err != nil {
			return c.Status(400).SendString(err.Error())
		}

		filter := bson.D{{Key: "_id", Value: result.InsertedID}}
		record := collection.FindOne(c.Context(), filter)
		cEmp := &Employee{}

		record.Decode(cEmp)
		return c.Status(200).JSON(cEmp)
	})

	employee.Put("/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		empId, err := primitive.ObjectIDFromHex(id)

		if err != nil {
			return c.Status(400).JSON(err)
		}

		var employee Employee
		if err := c.BodyParser(employee); err != nil {
			return c.Status(400).SendString(err.Error())
		}

		query := bson.D{
			{Key: "_id", Value: empId},
		}

		update := bson.D{
			{
				Key: "$set", Value: bson.D{
					{Key: "name", Value: employee.Name},
					{Key: "salary", Value: employee.Salary},
					{Key: "age", Value: employee.Age},
				},
			},
		}

		err = mg.Db.Collection("employee").FindOneAndUpdate(c.Context(), query, update).Err()

		if err != nil {
			if err == mongo.ErrNilDocument {
				return c.Status(400).JSON(err)
			}
			return c.Status(500).JSON(err)
		}

		employee.ID = id
		return c.JSON(employee)
	})

	employee.Get("/:id", func(c *fiber.Ctx) error {

		id := c.Params("id")
		empId, err := primitive.ObjectIDFromHex(id)

		if err != nil {
			return c.Status(400).JSON(err)
		}

		query := bson.D{{Key: "_id", Value: empId}}

		record := mg.Db.Collection("employee").FindOne(c.Context(), query)
		cEmp := &Employee{}

		record.Decode(cEmp)
		return c.Status(200).JSON(cEmp)
	})

	employee.Delete("/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		empId, err := primitive.ObjectIDFromHex(id)

		if err != nil {
			return c.Status(400).JSON(err)
		}

		query := bson.D{{Key: "_id", Value: empId}}

		result, err := mg.Db.Collection("employee").DeleteOne(c.Context(), &query)

		if err != nil {
			return c.Status(500).JSON(err)
		}

		if result.DeletedCount < 0 {
			return c.Status(404).JSON(err)
		}

		return c.JSON("record has been deleted")
	})

	app.Listen(":3000")
}
