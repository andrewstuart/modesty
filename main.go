package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func checkSuffix(s string, suffs ...string) bool {
	for _, suff := range suffs {
		if strings.HasSuffix(s, suff) {
			return true
		}
	}
	return false
}

func main() {
	allowed := []string{"github.com"}
	urls := os.Getenv("MODESTY_URLS")
	if urls != "" {
		for _, str := range strings.Split(urls, ",") {
			allowed = append(allowed, str)
		}
	}

	f, err := os.OpenFile("go.mod", os.O_RDONLY, 0400)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	br := bufio.NewReader(f)

	for {
		bs, _, err := br.ReadLine()
		if err != nil {
			break
		}
		fs := strings.Fields(string(bs))
		if len(fs) < 2 || fs[0] == "replace" {
			continue
		}
		u, err := url.Parse("https://" + fs[0])
		if err != nil || !strings.Contains(u.Host, ".") || checkSuffix(u.Host, allowed...) {
			continue
		}

		q := u.Query()
		q.Add("go-mod", "1")
		u.RawQuery = q.Encode()

		res, err := http.Get(u.String())
		if err != nil {
			continue
		}

		doc, err := goquery.NewDocumentFromReader(res.Body)
		if err != nil {
			continue
		}
		res.Body.Close()

		c, _ := doc.Find("meta[name=\"go-source\"]").Attr("content")
		contFs := strings.Fields(c)
		if len(contFs) < 2 {
			continue
		}
		u, err = url.Parse(contFs[1])
		if err != nil {
			continue
		}
		fmt.Printf("replace %s => %s/%s %s\n\n", fs[0], u.Host, strings.Trim(u.Path, "/"), fs[1])
	}
}
