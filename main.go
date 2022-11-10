package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/cli/go-gh"
	"github.com/cli/go-gh/pkg/repository"
)

func main() {
	repoOverride := flag.String("repo", "", "Specify a repository. If omitted, uses current repository")
	flag.Parse()

	var repo repository.Repository
	var err error

	if *repoOverride == "" {
		repo, err = gh.CurrentRepository()
	} else {
		repo, err = repository.Parse(*repoOverride)
	}
	if err != nil {
		fmt.Printf("could not determine what repo to use: %s\n", err.Error())
		os.Exit(1)
	}

	fmt.Printf("Going to search discussions in %s/%s\n", repo.Owner(), repo.Name())

	// TODO parse search arguments
	// TODO talk to API
	// TODO print results
}
