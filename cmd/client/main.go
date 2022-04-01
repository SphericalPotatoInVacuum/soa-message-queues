package main

import (
	"context"
	"fmt"
	"time"

	"github.com/Songmu/prompter"
	"github.com/SphericalPotatoInVacuum/soa-message-queues/internal/utils"
	pb "github.com/SphericalPotatoInVacuum/soa-message-queues/proto_gen/pathfinder"
	"github.com/briandowns/spinner"
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
	s := spinner.New(spinner.CharSets[13], 100*time.Millisecond)
	s.Reverse()
	s.Suffix = fmt.Sprintf(" Connecting to %s", addr)
	s.FinalMSG = fmt.Sprintf("Connected to %s!\n", addr)
	s.Start()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	conn, err := grpc.DialContext(ctx, addr, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		s.FinalMSG = "😭 Could not connect!\n"
		s.Stop()
	}
	utils.FailOnError(err, "Connection timed out")
	s.Stop()

	return &Client{
		grpcStub: pb.NewPathfinderClient(conn),
		conn:     conn,
	}
}

func main() {
	addr := prompter.Prompt("Enter pathfinder server address: ", "server:51075")
	client := NewClient(addr)
	defer client.Destroy()

	url1 := prompter.Prompt("Enter first URL", "https://en.wikipedia.org/wiki/Logitech")
	url2 := prompter.Prompt("Enter second URL", "https://en.wikipedia.org/wiki/Monsoon")

	s := spinner.New(spinner.CharSets[13], 100*time.Millisecond)
	s.Reverse()
	s.Suffix = "Finding a path..."
	s.FinalMSG = "Found path: "

	s.Start()
	resp, err := client.Find(context.Background(), &pb.FindRequest{URL1: url1, URL2: url2})
	utils.FailOnError(err, "Could not get result")
	s.Stop()
	fmt.Println(resp.String())
}
