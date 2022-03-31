package graphservice

import (
	"sync"

	"github.com/SphericalPotatoInVacuum/soa-message-queues/internal/queue"
	"github.com/SphericalPotatoInVacuum/soa-message-queues/internal/utils"
	grabber_pb "github.com/SphericalPotatoInVacuum/soa-message-queues/proto_gen/grabber"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"google.golang.org/protobuf/proto"
)

var failOnError = utils.FailOnError

type graphNode struct {
	Neighbors  []string
	Discovered bool
	Mu         sync.Mutex
}

type GraphService struct {
	// graph
	mu    sync.RWMutex
	graph map[string]*graphNode

	// rabbitmq
	conn *queue.Connection

	// requests
	waiting *WaitingMap
}

func NewGraphService(addr string) *GraphService {
	return &GraphService{
		graph:   make(map[string]*graphNode),
		conn:    queue.NewConnection(addr),
		waiting: NewWaitingMap(),
	}
}

func (g *GraphService) Start(num int) {
	consumer := g.conn.NewResultConsumer()
	for i := 0; i < num; i++ {
		go g.receive(consumer)
	}
}

func (g *GraphService) Destroy() {
	g.conn.Destroy()
}

func (g *GraphService) GetNeighbors(pageUrl string) []string {
	// get node from graph
	g.mu.RLock()
	node, exists := g.graph[pageUrl]
	if exists && node.Discovered {
		g.mu.RUnlock()
		return node.Neighbors
	}
	g.mu.RUnlock()

	g.mu.Lock()
	node = &graphNode{}
	node.Mu.Lock()
	defer node.Mu.Unlock()
	g.graph[pageUrl] = node
	g.mu.Unlock()

	reqId := g.discover(pageUrl)
	node.Discovered = true
	ch, _ := g.waiting.Get(reqId)
	node.Neighbors = <-ch

	return node.Neighbors
}

func (g *GraphService) discover(pageUrl string) string {
	reqId := uuid.NewString()
	req := grabber_pb.GrabRequest{
		RequestID: reqId,
		URL:       pageUrl,
	}
	body, err := proto.Marshal(&req)
	if err != nil {
		failOnError(err, "Failed marshal")
	}
	g.conn.GrabberProduce(body)
	log.WithField("reqID", req.RequestID).Info("Put task to queue")

	ch := make(chan []string)
	g.waiting.Put(reqId, ch)

	return reqId
}

func (g *GraphService) receive(consumer <-chan amqp.Delivery) {
	for msg := range consumer {
		resp := grabber_pb.GrabResponse{}
		proto.Unmarshal(msg.Body, &resp)

		log.WithField("reqID", resp.RequestID).Info("Got results from queue")
		ch, _ := g.waiting.Get(resp.RequestID)
		ch <- resp.URLs
	}
}
