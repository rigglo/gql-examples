package main

import (
	"log"
	"net/http"
	"time"

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
			"server_time": &gql.Field{
				Type: gql.Int,
				Resolver: func(c gql.Context) (interface{}, error) {
					out := make(chan interface{})
					go func() {
						ch := pinger()
						for {
							select {
							case <-c.Context().Done():
								// close some connections here
								log.Println("done")
								return
							case t := <-ch:
								log.Println("sending server time")
								out <- t
							}
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

func pinger() chan interface{} {
	ch := make(chan interface{})
	go func() {
		for {
			time.Sleep(2 * time.Second)
			ch <- time.Now().Unix()
		}
	}()
	return ch
}
