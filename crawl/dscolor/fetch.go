// Copyright (C) 2018 Ramesh Vyaghrapuri. All rights reserved.
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

// main can be called like so:
//
//   go run fetch.go http://danielsmith.com/category/wc15ml/ > ds_watercolors.json
//   go run fetch.go http://danielsmith.com/category/ooil/ > ds_oil.json
//
package main

import (
	"encoding/json"
	"fmt"
	"golang.org/x/net/html"
	"image"
	_ "image/jpeg"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
)

var nameSplitter = regexp.MustCompile(" [0-9]+ml.*")

func main() {
	processProductListPage(os.Args[1])
}

func processProductListPage(url string) {
	pages := []string{url}

	processPage(url, func(n *html.Node) {
		recurseNode(n, func(n *html.Node) bool {
			if getAttribute(n, "class") == "page" {
				pages = append(pages, getAttribute(n, "href"))
			}
			return false
		})
	})

	results := []map[string]string{}
	for _, page := range pages {
		processPage(page, func(n *html.Node) {
			log.Println("Processing", page)
			recurseNode(n, func(n *html.Node) bool {
				if getAttribute(n, "class") == "post_title" {
					href := getAttribute(n.FirstChild, "href")
					log.Println("Got product url", href)
					text := n.FirstChild.FirstChild.Data
					text = nameSplitter.ReplaceAllString(text, "")
					result := processProductPage(href)
					result["Name"] = text
					results = append(results, result)
				}
				return false
			})
		})
	}

	b, _ := json.MarshalIndent(results, "", "  ")
	fmt.Println(string(b))
}

func processProductPage(url string) map[string]string {
	var result map[string]string
	isOil := strings.Contains(url, "oil")

	processPage(url, func(n *html.Node) {
		recurseNode(n, func(n *html.Node) bool {
			if getAttribute(n, "class") == "the_content_wrapper" {
				result = getProductInfo(n.FirstChild.NextSibling.NextSibling)
				return true
			}
			return false
		})
		if imageURL := getImageURL(n); imageURL != "" {
			result["Colors"] = getImageColors(imageURL, isOil)
		}
	})
	result["Url"] = url
	return result
}

func getProductInfo(n *html.Node) map[string]string {
	text := ""
	recurseNode(n, func(n *html.Node) bool {
		if n.Type == html.TextNode {
			text += n.Data
		}
		return false
	})
	pairs := strings.Split(strings.Replace(text, " | ", "\n", 1), "\n")
	result := map[string]string{}
	for _, pair := range pairs {
		keyval := strings.Split(pair, ": ")
		if len(keyval) == 2 {
			result[keyval[0]] = keyval[1]
		}
	}
	return result
}

func getImageColors(url string, isOil bool) string {
	response, err := http.Get(url)
	if err != nil {
		fmt.Println("Failed to fetch", url, err)
		return "unknown"
	}

	defer func() { _ = response.Body.Close() }()
	m, _, err := image.Decode(response.Body)
	if err != nil {
		fmt.Println("Failed to decode", err)
		return "unknown"
	}

	var result []string
	if isOil {
		result = []string{
			sampleColor(m, 300, 100, 20),
			sampleColor(m, 300, 300, 20),
		}
	} else {
		result = []string{
			sampleColor(m, 300, 100, 10),
			sampleColor(m, 300, 245, 10),
			sampleColor(m, 300, 280, 10),
			sampleColor(m, 300, 310, 10),
			sampleColor(m, 300, 340, 10),
			sampleColor(m, 300, 375, 10),
		}
	}
	return strings.Join(result, " ")
}

func sampleColor(m image.Image, x, y, diff int) string {
	var r, g, b, count uint32
	for xx := x - diff; xx < x+diff; xx++ {
		for yy := y - diff; yy < y+diff; yy++ {
			r1, g1, b1, _ := m.At(xx, yy).RGBA()
			r += r1
			g += g1
			b += b1
			count++
		}
	}
	return fmt.Sprintf("%d,%d,%d", r/count/255, g/count/255, b/count/255)
}

func getImageURL(n *html.Node) string {
	result := ""
	recurseNode(n, func(n *html.Node) bool {
		if n.Type == html.ElementNode && n.Data == "meta" && getAttribute(n, "property") == "og:image" {
			result = getAttribute(n, "content")
			return true
		}
		return false
	})
	return result
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
