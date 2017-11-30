package main

import (
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// post is the content of a requested url
type post struct {
	Title           string    // is the file name
	When            time.Time // last modified time
	ContentFilePath string    // content of the file
	Tags            []string  // most common terms in the content, can be empty
}

type blog struct {
	sync.Mutex
	Posts map[string]post
}

var myblog blog

func main() {
	if len(os.Args) < 2 {
		log.Fatalln("need to provide a path to blog content")
	}

	path := os.Args[1]
	// populate stuff in blog
	myblog.Posts = make(map[string]post)
	err := filepath.Walk(path, func(lpath string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			p := post{
				Title:           info.Name(),
				When:            info.ModTime(),
				ContentFilePath: lpath,
			}
			postPath, _ := filepath.Rel(path, lpath)
			myblog.Posts[postPath] = p
		}
		return nil
	})
	if err != nil {
		log.Fatalln(err)
	}

	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}

func handler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path[1:]
	p, ok := myblog.Posts[path]
	if !ok {
		// 404
		http.NotFound(w, r)
		return
	}
	t := template.New("post")
	t, err := t.Parse(`<html>
<head>
    <title>{{ .Title }}</title>
</head>
<body>
    <h1>{{ .Title }}</h1>
    <h3>{{ .When }}</h3>
    {{ .Body }}
</body>
</html>`)

	if err != nil {
		log.Println(err)
	}

	b, err := ioutil.ReadFile(p.ContentFilePath)
	if err != nil {
		log.Println(err)
	}

	err = t.Execute(w, struct {
		Title string
		When  time.Time
		Body  template.HTML
	}{
		Title: p.Title,
		When:  p.When,
		Body:  template.HTML(b),
	})

	if err != nil {
		log.Println(err)
	}
}
