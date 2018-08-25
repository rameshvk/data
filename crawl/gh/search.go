// Copyright (C) 2018 Ramesh Vyaghrapuri. All rights reserved.
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

// go run search.go 'https://github.com/search?o=desc&q=language%3Ago&s=updated&type=Repositories'
package main

import (
	"encoding/json"
	"fmt"
	"golang.org/x/net/html"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	repos := []Repo(nil)
	url := os.Args[1]
	for count := 0; count < 5 && url != ""; count++ {
		r, n := fetchRecent(os.Args[1])
		repos = append(repos, r...)
		url = n
	}

	b, _ := json.MarshalIndent(repos, "", "  ")
	fmt.Println(string(b))
}

type Repo struct {
	Name, Description string
	LastUpdated       time.Time
	StarGazersCount   int
}

func fetchRecent(url string) ([]Repo, string) {
	repos := []Repo{}
	next := ""
	processPage(url, func(n *html.Node) {
		recurseNode(n, func(n *html.Node) bool {
			if name := getRepoName(n); name != "" {
				repos = append(repos, Repo{Name: name})
			}
			if text := getRepoDescription(n); text != "" {
				last := len(repos) - 1
				if strings.Contains(text, "Updated") {
					repos[last].LastUpdated = getRepoUpdated(n)
				} else if repos[last].Description == "" && !strings.Contains(text, "icense") {
					repos[last].Description = text
				}
			}
			if n := getRepoStargazers(n); n > 0 {
				repos[len(repos)-1].StarGazersCount = n
			}

			if text := getRepoNextLink(url, n); text != "" {
				next = text
			}

			return false
		})
	})
	return repos, next
}

func getRepoNextLink(base string, n *html.Node) string {
	if n.Type == html.ElementNode && n.Data == "a" && getAttribute(n, "rel") == "next" {
		baseUri, err1 := url.Parse(base)
		relUri, err2 := url.Parse(getAttribute(n, "href"))
		if err1 == nil && err2 == nil {
			return baseUri.ResolveReference(relUri).String()
		} else {
			log.Println("Unexpected URL parse errors", err1, err2)
		}
	}
	return ""
}

func getRepoUpdated(n *html.Node) time.Time {
	attr := getAttribute(n.FirstChild.NextSibling, "datetime")
	t, err := time.Parse("2006-01-02T15:04:05Z", attr)
	if err != nil {
		panic(err)
	}
	return t.Local()
}

func getRepoDescription(n *html.Node) string {
	if n.Type == html.ElementNode && n.Data == "p" && strings.Contains(getAttribute(n, "class"), "text-gray") {
		return strings.TrimSpace(innerText(n))
	}
	return ""
}

func getRepoStargazers(n *html.Node) int {
	if n.Type != html.ElementNode || n.Data != "svg" {
		return 0
	}
	if getAttribute(n, "aria-label") != "star" {
		return 0
	}
	if nn, err := strconv.Atoi(strings.TrimSpace(innerText(n.NextSibling))); err == nil {
		return nn
	} else {
		log.Println("Unexpected stargazer", err, strings.TrimSpace(innerText(n)))
	}

	return 0
}

func getRepoName(n *html.Node) string {
	if n.Type != html.ElementNode || n.Data != "a" {
		return ""
	}

	if getAttribute(n, "class") != "v-align-middle" {
		return ""
	}

	if n.FirstChild == nil || n.FirstChild.Type != html.TextNode {
		log.Println("Unexpected node", n.FirstChild)
		return ""
	}

	href := getAttribute(n, "href")
	if "/"+n.FirstChild.Data != href {
		log.Println("Unexpected node data", n.FirstChild.Data)
		return ""
	}

	return n.FirstChild.Data
}

func innerText(n *html.Node) string {
	text := ""
	recurseNode(n, func(n *html.Node) bool {
		if n.Type == html.TextNode {
			text += n.Data
		}
		return false
	})
	return text
}

func getAttribute(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}

func recurseNode(n *html.Node, fn func(n *html.Node) bool) bool {
	if fn(n) {
		return true
	}
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		if recurseNode(child, fn) {
			return true
		}
	}
	return false
}

func processPage(url string, fn func(n *html.Node)) {
	response, err := http.Get(url)
	if err != nil {
		fmt.Println("Failed to fetch", url, err)
		return
	}

	defer func() { _ = response.Body.Close() }()
	z, err := html.Parse(response.Body)
	if err != nil {
		fmt.Println("Error parsing body", err)
		return
	}
	fn(z)
}
