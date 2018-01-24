package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/html"
)

func fetch() (*html.Node, error) {
	jar, _ := cookiejar.New(nil)
	randoUrl, _ := url.Parse("http://www.korearandonneurs.kr")

	cookie := &http.Cookie{
		Name:     "myname",
		Value:    "myvalue",
		Unparsed: []string{"lang=ko", "PHPSESSID=u0jsf5rvmo6fcip9e1mkljq3h6"},
	}

	jar.SetCookies(randoUrl, []*http.Cookie{cookie})

	client := &http.Client{
		Jar: jar,
	}

	postData := url.Values{}
	postData.Set("email", "aiceru@gmail.com")
	postData.Set("member_num", "12659")
	postData.Set("target", "register.php")

	req, err := http.NewRequest("POST", "http://www.korearandonneurs.kr/reg/login_do.php", strings.NewReader(postData.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	if err != nil {
		log.Println(err)
		return nil, err
	}

	res, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	doc, err := html.Parse(res.Body)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return doc, nil
}

func parse(doc *html.Node) bool {
	var f func(*html.Node) bool
	f = func(n *html.Node) bool {
		if n.Type == html.TextNode && n.Data == "Busan 200K" {
			eventNode := n.Parent.Parent.NextSibling.FirstChild

			if eventNode.Data == "div" && eventNode.Attr[0].Val == "event-descr" &&
				strings.Contains(eventNode.FirstChild.Data, "Fee: Please refer to") {
				fmt.Println(n.Data)
				return true
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if f(c) {
				return true
			}
		}
		return false
	}
	return f(doc)
}

func renderNode(n *html.Node) string {
	var buf bytes.Buffer
	w := io.Writer(&buf)
	html.Render(w, n)
	return buf.String()
}

func main() {
	for {
		doc, err := fetch()
		if err != nil {
			log.Println(err)
			return
		}

		if parse(doc) == true {
			fmt.Println("SEND MAIL")
		}
		time.Sleep(2 * time.Second)
	}
}
