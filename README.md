# authing-go-sdk

## What is Authing

[Authing](https://authing.cn/) is an IDaaS which is created by Ivy.

## Installation
Make sure you have a working Go environment. To install `authing-go-sdk`, simply run:

```shell
go get github.com/Authing/graphql
```

## Quick Guide

```go
package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/Authing/authing-go-sdk"
	"github.com/kelvinji2009/graphql"
)

const (
	clientID  = "5adb75e03055230001023b26"
	appSecret = "e683d18f9d597317d43d7a6522615b9d"
)

func main() {
	client := authing.NewClient(clientID, appSecret, false)
	// Enable debug info for graphql client, just comment it if you want to disable the debug info
	client.Client.Log = func(s string) { log.Println(s) }

	// >>>Graphql Mutation: register
	input := authing.UserRegisterInput{
		Email:            graphql.String("kelvinji2009@gmail.com"),
		Password:         graphql.String("password"),
		RegisterInClient: graphql.String(clientID),
	}

	m, err := client.Register(&input)
	if err != nil {
		log.Println(">>>>Register failed: " + err.Error())
	} else {
		printJSON(m)
	}

	// oauthClient := authing.NewOauthClient(clientID, appSecret, false)
	// Enable debug info for graphql client, just comment it if you want to disable the debug info
	// oauthClient.Client.Log = func(s string) { log.Println(s) }

	// >>>>Graphql Query: Read OAuth List
	// readOauthListQueryParameter := authing.ReadOauthListQueryParameter{
	// 	ClientID:   graphql.String(clientID),
	// 	DontGetURL: graphql.Boolean(false),
	// }

	// q, err := oauthClient.ReadOauthList(&readOauthListQueryParameter)
	// if err != nil {
	// 	log.Println(">>>>Read OAuth List failed: " + err.Error())
	// } else {
	// 	printJSON(q)
	// }

}

// printJSON prints v as JSON encoded with indent to stdout. It panics on any error.
func printJSON(v interface{}) {
	w := json.NewEncoder(os.Stdout)
	w.SetIndent("", "\t")
	err := w.Encode(v)
	if err != nil {
		panic(err)
	}
}
```

## TODO

- [ ] More detailed API usages and documents
- [ ] Travis CI support


## Thanks
[Go GraphQL Client](https://github.com/shurcooL/graphql)

[Simple low-level GraphQL HTTP client for Go](https://github.com/machinebox/graphql)

