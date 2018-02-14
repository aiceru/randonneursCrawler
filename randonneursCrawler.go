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

func login(cli *http.Client) {
	postData := url.Values{}
	postData.Set("email", "aiceru@gmail.com")
	postData.Set("member_num", "12659")
	postData.Set("target", "register.php")

	req, err := http.NewRequest("POST", "http://www.korearandonneurs.kr/reg/login_do.php", strings.NewReader(postData.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Connnection", "close")
	if err != nil {
		log.Println(err)
		return
	}

	_, err = cli.Do(req)
	for err != nil {
		log.Println(err, ", retrying...")
		time.Sleep(1 * time.Second)
		_, err = cli.Do(req)
	}
}

func register(cli *http.Client, bid string) {
	postData := url.Values{}
	postData.Set("event_id", bid)
	req, err := http.NewRequest("POST", "http://www.korearandonneurs.kr/reg/event_apply.php", strings.NewReader(postData.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Connection", "close")
	if err != nil {
		log.Println(err)
		return
	}

	res, err := cli.Do(req)
	if err != nil {
		log.Println(err)
		return
	}
	defer res.Body.Close()

	doc, err := html.Parse(res.Body)
	fmt.Println(renderNode(doc))
}

func fetch(cli *http.Client) (*html.Node, error) {
	req, err := http.NewRequest("GET", "http://www.korearandonneurs.kr/reg/register.php", nil)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Connection", "close")
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
	defer res.Body.Close()

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

func parse(doc *html.Node) (bool, string) {
	var id string
	var f func(*html.Node) (bool, string)
	f = func(n *html.Node) (bool, string) {
		if n.Type == html.TextNode {
			for i := range brevets {
				brevet := &(brevets[i])
				if n.Data == "Register" {
					eventNode := n.Parent.Parent.PrevSibling
					if eventNode.FirstChild.FirstChild.Data == brevet.date && eventNode.LastChild.FirstChild.Data == brevet.name {
						id = n.Parent.Attr[1].Val
						return true, id
					}
				}
				/*if n.Data == brevet.name && n.Parent.PrevSibling.FirstChild.Data == brevet.date {
					eventNode := n.Parent.Parent.NextSibling.FirstChild

					if eventNode.Data == "div" && eventNode.Attr[0].Val == "event-descr" &&
						strings.Contains(eventNode.FirstChild.Data, "Fee: Please refer to") {
						//brevetId := eventNode.NextSibling.Attr[1].Val
						fmt.Println(brevet.name)
						logger.Println(brevet.name + ".avail goes true")
						brevet.avail = true
					} else {
						logger.Println(brevet.name + ".avail goes false")
						brevet.avail = false
					}
				}*/
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if success, id := f(c); success {
				return true, id
			}
		}
		return false, ""
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
	fpLog, err := os.OpenFile("rando_log.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Panic(err)
	}
	defer fpLog.Close()

	logger = log.New(fpLog, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)

	brevets = []Brevet{
		Brevet{"Seoul 200K", "10 Mar Sat", false, true},
		//Brevet{"Seoul 300K", "31 Mar Sat", false, true},
		//Brevet{"Seoul 400K", "21 Apr Sat", false, true},
		//Brevet{"Busan 200K", "3 Mar Sat", false, true},
	}

	jar, _ := cookiejar.New(nil)
	randoUrl, _ := url.Parse("http://www.korearandonneurs.kr")

	cookie := &http.Cookie{
		Name:  "lang",
		Value: "en",
	}

	client := &http.Client{
		Jar: jar,
	}

	jar.SetCookies(randoUrl, []*http.Cookie{cookie})

	for {
		login(client)

		doc, err := fetch(client)
		if err != nil {
			log.Println(err)
			return
		}

		if success, brevetId := parse(doc); success {
			register(client, brevetId)
			log.Println("successfully registered")
			os.Exit(0)
		}
		time.Sleep(500 * time.Millisecond)
	}

	/*for {
		doc, err := fetch()
		if err != nil {
			log.Println(err)
			return
		}

		parse(doc)
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
