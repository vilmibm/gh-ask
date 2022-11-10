package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/cli/go-gh"
	"github.com/cli/go-gh/pkg/repository"
)

func main() {
	repoOverride := flag.String(
		"repo", "",
		"Specify a repository. If omitted, uses current repository")
	flag.Parse()

	var repo repository.Repository
	var err error

	if *repoOverride == "" {
		repo, err = gh.CurrentRepository()
	} else {
		repo, err = repository.Parse(*repoOverride)
	}
	if err != nil {
		fmt.Printf(
			"could not determine what repo to use: %s\n", err.Error())
		os.Exit(1)
	}

	if len(os.Args) < 2 {
		fmt.Println("Please specify a search term")
		os.Exit(2)
	}
	search := strings.Join(os.Args[1:], " ")

	fmt.Printf(
		"Going to search discussions in '%s/%s' for '%s'\n",
		repo.Owner(), repo.Name(), search)

	// TODO talk to API
	// TODO print results
}
