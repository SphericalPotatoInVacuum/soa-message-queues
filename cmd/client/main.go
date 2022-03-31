package main

import (
	"bufio"
	"context"
	"os"
	"strings"
	"time"

	"github.com/SphericalPotatoInVacuum/soa-message-queues/internal/utils"
	pb "github.com/SphericalPotatoInVacuum/soa-message-queues/proto_gen/pathfinder"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

var getEnv = utils.GetEnv

type grpcStub = pb.PathfinderClient
type Client struct {
	grpcStub

	conn *grpc.ClientConn
}

func (c *Client) Destroy() {
	c.conn.Close()
}

func NewClient(addr string) *Client {
	log.Infof("Trying to connect to server %s", addr)
	conn, err := grpc.Dial(addr, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(time.Duration(30)*time.Second))
	if err != nil {
		log.Fatalf("Could not connect: %v", err)
	}
	log.Infof("Connected to server: %s", addr)

	return &Client{
		grpcStub: pb.NewPathfinderClient(conn),
		conn:     conn,
	}
}

func main() {
	log.SetFormatter(&log.TextFormatter{
		ForceColors: true,
	})

	addr := getEnv("PATHFINDER_SERVER_ADDR", "127.0.0.1:51075")
	client := NewClient(addr)
	defer client.Destroy()

	reader := bufio.NewReader(os.Stdin)

	url1, err := reader.ReadString('\n')
	utils.FailOnError(err, "Could not read string")
	url1 = strings.TrimSpace(url1)

	url2, err := reader.ReadString('\n')
	utils.FailOnError(err, "Could not read string")
	url2 = strings.TrimSpace(url2)

	resp, err := client.Find(context.Background(), &pb.FindRequest{URL1: url1, URL2: url2})
	if err != nil {
		log.Fatal(err)
	}
	log.Info(resp.String())
}
