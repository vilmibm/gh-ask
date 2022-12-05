package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/cli/go-gh"
	"github.com/cli/go-gh/pkg/browser"
	"github.com/cli/go-gh/pkg/jq"
	"github.com/cli/go-gh/pkg/jsonpretty"
	"github.com/cli/go-gh/pkg/repository"
	"github.com/cli/go-gh/pkg/tableprinter"
	"github.com/cli/go-gh/pkg/term"
)

func main() {
	jsonFlag := flag.Bool("json", false, "Output JSON")
	jqFlag := flag.String("jq", "", "Process JSON output with a jq expression")
	lucky := flag.Bool("lucky", false, "Open the first matching result in a web browser")
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

	if len(flag.Args()) < 1 {
		fmt.Println("Please specify a search term")
		os.Exit(2)
	}
	search := strings.Join(flag.Args(), " ")

	client, err := gh.GQLClient(nil)
	if err != nil {
		fmt.Printf("could not create a graphql client: %s", err.Error())
		os.Exit(3)
	}

	query := fmt.Sprintf(`{
			repository(owner: "%s", name: "%s") {
				hasDiscussionsEnabled
				discussions(first: 100) {
					edges { node {
						title
						body
						url
      }}}}}`, repo.Owner(), repo.Name())

	type Discussion struct {
		Title string
		URL   string `json:"url"`
		Body  string
	}

	response := struct {
		Repository struct {
			Discussions struct {
				Edges []struct {
					Node Discussion
				}
			}
			HasDiscussionsEnabled bool
		}
	}{}

	err = client.Do(query, nil, &response)
	if err != nil {
		fmt.Printf("failed to talk to the GitHub API: %s", err.Error())
		os.Exit(4)
	}

	if !response.Repository.HasDiscussionsEnabled {
		fmt.Printf("%s/%s does not have discussions enabled.\n", repo.Owner(), repo.Name())
		os.Exit(5)
	}

	matches := []Discussion{}

	for _, edge := range response.Repository.Discussions.Edges {
		if strings.Contains(edge.Node.Body+edge.Node.Title, search) {
			matches = append(matches, edge.Node)
		}
	}

	if len(matches) == 0 {
		fmt.Println("No matching discussion threads found :(")
	}

	if *lucky {
		b := browser.New("", os.Stdout, os.Stderr)
		b.Browse(matches[0].URL)
		return
	}

	isTerminal := term.IsTerminal(os.Stdout)

	if *jsonFlag {
		output, err := json.Marshal(matches)
		if err != nil {
			fmt.Printf("could not serialize JSON: %s", err.Error())
			os.Exit(7)
		}
		if *jqFlag != "" {
			err = jq.Evaluate(bytes.NewBuffer(output), os.Stdout, *jqFlag)
			if err != nil {
				fmt.Printf("failed to execute jq: %s", err.Error())
				os.Exit(8)
			}
			return
		}
		err = jsonpretty.Format(os.Stdout, bytes.NewBuffer(output), " ", isTerminal)
		if err != nil {
			fmt.Printf("could not format JSON: %s", err.Error())
			os.Exit(9)
		}

		return
	}

	tp := tableprinter.New(os.Stdout, isTerminal, 100)

	if isTerminal {
		fmt.Printf(
			"Searching discussions in '%s/%s' for '%s'\n",
			repo.Owner(), repo.Name(), search)
	}

	fmt.Println()
	for _, d := range matches {
		tp.AddField(d.Title)
		tp.AddField(d.URL)
		tp.EndRow()
	}

	err = tp.Render()
	if err != nil {
		fmt.Printf("could not render data: %s", err.Error())
		os.Exit(6)
	}
}
