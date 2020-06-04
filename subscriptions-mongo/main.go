package main

import (
	"fmt"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/rigglo/gql"
	"github.com/rigglo/gql/pkg/handler"
	"github.com/rigglo/gqlws"
)

func main() {
	exec := gql.DefaultExecutor(Schema)

	h := handler.New(handler.Config{
		Executor:   exec,
		Playground: true,
	})

	wsh := gqlws.New(
		gqlws.Config{
			Subscriber: exec.Subscribe,
		},
		h,
	)

	http.Handle("/graphql", wsh)
	if err := http.ListenAndServe(":9999", nil); err != nil {
		panic(err)
	}
}

var (
	RootSubscription = &gql.Object{
		Name: "Subscription",
		Fields: gql.Fields{
			"new_things": &gql.Field{
				Type: gql.String,
				Resolver: func(c gql.Context) (interface{}, error) {
					out := make(chan interface{})
					client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://foo:bar@localhost:27017"))
					if err != nil {
						return nil, err
					}
					err = client.Connect(c.Context())
					if err != nil {
						return nil, err
					}

					database := client.Database("foobar")
					productsCollection := database.Collection("products")

					/* matchPipeline := bson.D{
						{
							"$match", bson.D{
								{"operationType", "insert"},
							},
						},
					} */

					changeStream, err := productsCollection.Watch(c.Context(), mongo.Pipeline{})
					if err != nil {
						panic(err)
					}
					go func() {
						for changeStream.Next(c.Context()) {
							var data bson.M
							if err := changeStream.Decode(&data); err != nil {
								panic(err)
							}
							fmt.Printf("%v\n", data)
							out <- changeStream.Current.String()
						}
					}()
					return out, nil
				},
			},
		},
	}

	RootQuery = &gql.Object{
		Name:   "Query",
		Fields: gql.Fields{},
	}

	Schema = &gql.Schema{
		Query:        RootQuery,
		Subscription: RootSubscription,
	}
)

/*
use admin
db.createUser(
{
   user: "foo",
   pwd: "bar",
   roles: [ { role: "userAdminAnyDatabase", db: "admin" } ]
 }
)
*/
