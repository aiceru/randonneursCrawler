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
	client := &http.Client{Jar: jar}

	postData := strings.NewReader("email=aiceru@gmail.com&member_num=12659&target=register.php")

	req, err := http.NewRequest("POST", "http://www.korearandonneurs.kr/reg/login_do.php", postData)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	//req.AddCookie(&cookie)

	res, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	req, err = http.NewRequest("GET", "http://www.korearandonneurs.kr/reg/register.php", nil)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	//req.AddCookie(&cookie)

	res, err = client.Do(req)
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
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "img" {
		}
	}
	f(doc)

	return false
}

func renderNode(n *html.Node) string {
	var buf bytes.Buffer
	w := io.Writer(&buf)
	html.Render(w, n)
	return buf.String()
}

func main() {

	doc, err := fetch()
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Println(renderNode(doc))

	if parse(doc) == true {
		fmt.Println("SEND MAIL")
	}
}
