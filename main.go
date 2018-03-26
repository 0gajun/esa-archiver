package main

import (
	"context"
	"fmt"
	"os"

	"github.com/0gajun/esa-archiver/esa"
)

func main() {
	accessToken, ok := os.LookupEnv("ESA_TOKEN")
	if !ok {
		fmt.Printf("Missing ESA_TOKEN environment variable")
		os.Exit(1)
	}

	esaClient, err := esa.NewEsa(accessToken, "0gajun")
	if err != nil {
		fmt.Printf("%v", err)
		os.Exit(1)
	}

	posts, err := esaClient.GetAllPosts(context.Background())
	if err != nil {
		fmt.Printf("%v", err)
		os.Exit(1)
	}

	for _, post := range posts {
		fmt.Printf("%s\n", post.Name)
	}
}
