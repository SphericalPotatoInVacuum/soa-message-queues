package grabber

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/SphericalPotatoInVacuum/soa-message-queues/internal/queue"
	grabber_pb "github.com/SphericalPotatoInVacuum/soa-message-queues/proto_gen/grabber"
	log "github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
)

var excludedPatterns = []string{
	"File:",
	"Talk:",
	"Category:",
	"Special:",
	"Wikipedia:",
	"Help:",
	"Portal:",
}

func checkHref(href string) bool {
	if !strings.HasPrefix(href, "/wiki/") {
		return false
	}

	href = strings.TrimPrefix(href, "/wiki/")

	for _, pattern := range excludedPatterns {
		if strings.HasPrefix(href, pattern) {
			return false
		}
	}

	return true
}

func grab(urlStr string) ([]string, error) {
	log.WithField(
		"url", urlStr,
	).Info("Grabbing url")

	u, err := url.Parse(urlStr)
	baseUrl := fmt.Sprintf("%s://%s", u.Scheme, u.Hostname())

	res, err := http.Get(u.String())
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	links := make([]string, 0)
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if exists && checkHref(href) {
			link := fmt.Sprintf("%s%s", baseUrl, href)
			links = append(links, link)
		}
	})

	return links, nil
}

type Grabber struct {
	// rabbitmq
	conn *queue.Connection
}

func NewGrabber(addr string) *Grabber {
	return &Grabber{
		conn: queue.NewConnection(addr),
	}
}

func (g *Grabber) Destroy() {
	g.conn.Destroy()
}

func (g *Grabber) Run() {
	msgs := g.conn.NewGrabberConsumer()
	for msg := range msgs {
		var req grabber_pb.GrabRequest
		proto.Unmarshal(msg.Body, &req)
		log.WithField("reqID", req.RequestID).Info("Got task from queue")

		links, err := grab(req.URL)
		if err != nil {
			msg.Nack(false, false)
			log.Error(err)
			continue
		}
		log.WithField(
			"url", req.URL,
		).Info("Grabbed url")
		body, err := proto.Marshal(&grabber_pb.GrabResponse{
			RequestID: req.RequestID,
			URLs:      links,
		})
		g.conn.ResultProduce(body)
		msg.Ack(false)
		log.WithField("reqID", req.RequestID).Info("Put results to queue")
	}
}
