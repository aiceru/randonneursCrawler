package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"
	"time"

	"golang.org/x/net/html"
)

var logger *log.Logger

func login(cli *http.Client) ([]*http.Cookie, error) {
	postData := url.Values{}
	postData.Set("email", "aiceru@gmail.com")
	postData.Set("member_num", "12659")
	//postData.Set("target", "register.php")

	req, err := http.NewRequest("POST", "http://www.korearandonneurs.kr/reg/login_do.php", strings.NewReader(postData.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	if err != nil {
		log.Println(err)
		return nil, err
	}

	res, err := cli.Do(req)
	for err != nil {
		log.Println(err, ", retrying...")
		time.Sleep(1 * time.Second)
		res, err = cli.Do(req)
	}

	fmt.Println(res.Cookies())

	return res.Cookies(), nil
}

func fetch() (*html.Node, error) {
	jar, _ := cookiejar.New(nil)
	randoUrl, _ := url.Parse("http://www.korearandonneurs.kr")

	/*cookie := &http.Cookie{
		//Name:  "myname",
		//Value: "myvalue",
		Unparsed: []string{"lang=en"},
	}*/

	client := &http.Client{
		Jar: jar,
	}

	c, err := login(client)
	jar.SetCookies(randoUrl, c)

	req, err := http.NewRequest("GET", "http://www.korearandonneurs.kr/reg/register.php", nil)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	if err != nil {
		log.Println(err)
		return nil, err
	}

	res, err := client.Do(req)
	for err != nil {
		log.Println(err, ", retrying...")
		time.Sleep(1 * time.Second)
		res, err = client.Do(req)
	}

	doc, err := html.Parse(res.Body)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return doc, nil
}

type Brevet struct {
	name  string
	date  string
	avail bool
	mail  bool
}

var brevets []Brevet

func parse(doc *html.Node) {
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.TextNode {
			for i := range brevets {
				brevet := &(brevets[i])
				if n.Data == brevet.name && n.Parent.PrevSibling.FirstChild.Data == brevet.date {
					eventNode := n.Parent.Parent.NextSibling.FirstChild

					if eventNode.Data == "div" && eventNode.Attr[0].Val == "event-descr" &&
						strings.Contains(eventNode.FirstChild.Data, "Fee: Please refer to") {
						logger.Println(brevet.name + ".avail goes true")
						brevet.avail = true
					} else {
						logger.Println(brevet.name + ".avail goes false")
						brevet.avail = false
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)
}

func renderNode(n *html.Node) string {
	var buf bytes.Buffer
	w := io.Writer(&buf)
	html.Render(w, n)
	return buf.String()
}

func main() {
	fpLog, err := os.OpenFile("rando_log.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Panic(err)
	}
	defer fpLog.Close()

	logger = log.New(fpLog, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)

	brevets = []Brevet{
		Brevet{"Seoul 200K", "10 Mar Sat", false, true},
		Brevet{"Seoul 300K", "31 Mar Sat", false, true},
		Brevet{"Seoul 400K", "21 Apr Sat", false, true},
	}

	//	for {
	_, err = fetch()
	//doc, err := fetch()
	if err != nil {
		log.Println(err)
		return
	}

	/*	parse(doc)
		for i := range brevets {
			brevet := &(brevets[i])
			if brevet.avail {
				if brevet.mail {
					mail := Mail{
						senderId: "js.pr.mailing",
						toIds: []string{
							"aiceru@gmail.com",
							"whiteamin@gmail.com",
							"dnjsdud0225@gmail.com",
							"aquanuri@gmail.com",
							"tlsrjsgk8987@naver.com",
							"genisus@naver.com",
						},
						subject: "Randonneurs register Noti",
						body: "Registering for randonnerus " + brevet.name + " at " + brevet.date + " is now available\n" +
							"Go and register now : http://www.korearandonneurs.kr/reg/login.php?target=register.php",
					}
					SendMail(mail)
					logger.Println("mail sended and " + brevet.name + ".mail goes false")
					brevet.mail = false
				}
			} else {
				logger.Println(brevet.name + ".mail goes true")
				brevet.mail = true
				logger.Println("Not found available event")
			}
		}
		time.Sleep(2 * time.Second)
		}*/
}
