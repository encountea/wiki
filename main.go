package main

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/gocolly/colly"
)

// Node представляет узел в пути поиска
type Node struct {
	url      string
	sentence string
	prev     *Node
}

// Извлечение ссылки и предложения из HTML-контента
func getLinks(url string) (map[string]string, error) {
	links := make(map[string]string)
	c := colly.NewCollector()

	c.OnHTML("p", func(e *colly.HTMLElement) {
		text := e.Text
		e.ForEach("a[href]", func(_ int, el *colly.HTMLElement) {
			href := el.Attr("href")
			if strings.HasPrefix(href, "/wiki/") && !strings.Contains(href, ":") {
				fullURL := "https://ru.wikipedia.org" + href
				if _, exists := links[fullURL]; !exists {
					links[fullURL] = getSentence(text, el.Text)
				}
			}
		})
	})

	err := c.Visit(url)
	if err != nil {
		return nil, err
	}

	return links, nil
}

// Извлечение предложения, содержащее ссылку
func getSentence(text, linkText string) string {
	re := regexp.MustCompile(`(?s)([^.!?]*?\(.*?\).*?[^.!?]*` + regexp.QuoteMeta(linkText) + `[^.!?]*?[.!?])|([^.!?]*` + regexp.QuoteMeta(linkText) + `[^.!?]*[.!?])`)
	matches := re.FindStringSubmatch(text)
	if len(matches) > 0 {
		if matches[1] != "" {
			return strings.TrimSpace(matches[1])
		}
		return strings.TrimSpace(matches[2])
	}
	return ""
}

// Нахождение пути от startURL до endURL
func findPath(startURL, endURL string, logger *log.Logger) ([]*Node, error) {
	queue := []*Node{{url: startURL}}
	visited := make(map[string]bool)
	visited[startURL] = true

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		logger.Println("Посещение:", current.url)

		if current.url == endURL {
			var path []*Node
			for node := current; node != nil; node = node.prev {
				path = append([]*Node{node}, path...)
			}
			if len(path) > 1 {
				path = path[1:]
			}
			return path, nil
		}

		links, err := getLinks(current.url)
		if err != nil {
			log.Println(err)
			continue
		}

		for link, sentence := range links {
			if !visited[link] {
				visited[link] = true
				queue = append(queue, &Node{url: link, sentence: sentence, prev: current})
			}
		}
	}

	return nil, fmt.Errorf("путь не найден")
}

func main() {
	var startURL, endURL string
	fmt.Println("Введите стартовую ссылку на страницу Wikipedia:")
	fmt.Scan(&startURL)
	fmt.Println("Введите конечную ссылку на страницу Wikipedia:")
	fmt.Scan(&endURL)

	logFile, err := os.OpenFile("log.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer logFile.Close()

	logger := log.New(logFile, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)

	path, err := findPath(startURL, endURL, logger)
	if err != nil {
		log.Fatalf("Ошибка при поиске пути: %v", err)
	}

	for i, node := range path {
		fmt.Printf("%d------------------------\n", i+1)
		fmt.Println(node.sentence)
		fmt.Println(node.url)
	}
}
