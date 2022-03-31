package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/SphericalPotatoInVacuum/soa-message-queues/internal/graphservice"
	"github.com/SphericalPotatoInVacuum/soa-message-queues/internal/safeset"
	"github.com/SphericalPotatoInVacuum/soa-message-queues/internal/utils"
	pb "github.com/SphericalPotatoInVacuum/soa-message-queues/proto_gen/pathfinder"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

var getEnv = utils.GetEnv
var failOnError = utils.FailOnError

type pathfinderServer struct {
	pb.UnimplementedPathfinderServer

	g *graphservice.GraphService
}

func (s *pathfinderServer) bfs(ctx context.Context, URL1 string, URL2 string) ([]string, error) {
	doneChan := make(chan []string)
	go func() {
		visited := safeset.NewSafeSet()
		paths := make(chan []string, 1000000)  // bfs queue
		buff := make(chan [][]string, 1000000) // intermediate buffer

		paths <- []string{URL1}
		breadth := 0
		// breadth loop
		for {
			// one level loop
			N := 0
		level_loop:
			for {
				var path []string
				select {
				case path = <-paths:
					break
				default:
					break level_loop
				}

				last := path[len(path)-1]
				go func() {
					visited.Insert(last)
					log.WithField("page", last).Info("Checking neighbors")
					newPaths := make([][]string, 0)
					for _, link := range s.g.GetNeighbors(path[len(path)-1]) {
						select {
						case <-ctx.Done():
							return
						default:
							break
						}

						if visited.Exists(link) {
							continue
						}
						newPath := make([]string, len(path), len(path)+1)
						copy(newPath, path)
						newPath = append(newPath, link)
						newPaths = append(newPaths, newPath)
					}

					buff <- newPaths
				}()

				N++

				select {
				case <-ctx.Done():
					return
				default:
					break
				}
			}

			log.WithFields(log.Fields{"breadth": breadth, "N": N}).Info("Started fetching for new breadth level")
			g, _ := errgroup.WithContext(ctx)
			for i := 0; i < N; i++ {
				g.Go(func() error {
					newPaths := <-buff
					for _, newPath := range newPaths {
						if newPath[len(newPath)-1] == URL2 {
							doneChan <- newPath
							return nil
						}
						paths <- newPath
					}
					return nil
				})
			}
			if err := g.Wait(); err != nil {
				log.Error(err)
				return
			}
		}
	}()

	select {
	case <-ctx.Done():
		log.Info("BFS timed out")
		return nil, errors.New("Request took too long")
	case path := <-doneChan:
		return path, nil
	}
}

func validateWikiUrl(urlStr string) error {
	u, err := url.Parse(urlStr)
	if err != nil {
		log.WithError(
			err,
		).Error("Parse url failed")
		return err
	}

	parts := strings.Split(u.Hostname(), ".")
	if len(parts) < 2 || parts[len(parts)-2] != "wikipedia" || parts[len(parts)-1] != "org" {
		err = fmt.Errorf("Expected wikipedia url, but received %s", urlStr)
		log.WithError(
			err,
		).Error("Invalid wikipedia url")
		return err
	}

	return nil
}

func (s *pathfinderServer) Find(ctx context.Context, in *pb.FindRequest) (*pb.FindResponse, error) {
	contextLogger := log.WithFields(log.Fields{
		"URL1": in.URL1,
		"URL2": in.URL2,
	})
	contextLogger.Infof("Processing request")
	if err := validateWikiUrl(in.URL1); err != nil {
		contextLogger.WithError(err).Info("URL1 validation failed")
		return nil, err
	}
	if err := validateWikiUrl(in.URL2); err != nil {
		contextLogger.WithError(err).Info("URL2 validation failed")
		return nil, err
	}
	if in.URL1 == in.URL2 {
		contextLogger.Info("Finished processing request")
		return &pb.FindResponse{
			Length: 0,
			Path:   []string{in.URL1},
		}, nil
	}

	ctx, cancel := context.WithTimeout(ctx, time.Duration(120)*time.Second)
	defer cancel()
	path, err := s.bfs(ctx, in.URL1, in.URL2)
	if err != nil {
		return nil, err
	}

	contextLogger.Info("Finished processing request")
	return &pb.FindResponse{
		Length: int32(len(path) - 1),
		Path:   path,
	}, nil
}

func newServer(graph *graphservice.GraphService) *pathfinderServer {
	s := pathfinderServer{
		g: graph,
	}
	return &s
}

func serve(graph *graphservice.GraphService) {
	port := os.Getenv("PATHFINDER_SERVER_PORT")
	if port == "" {
		port = "51075"
	}
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Infof("listening on port: %s", lis.Addr())
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterPathfinderServer(grpcServer, newServer(graph))
	grpcServer.Serve(lis)
}

func getRabbitAddr() string {
	user := getEnv("RABBITMQ_USER", "guest")
	pass := getEnv("RABBITMQ_PASS", "guest")
	host := getEnv("RABBITMQ_HOST", "127.0.0.1")
	port := getEnv("RABBITMQ_PORT", "5672")
	return fmt.Sprintf("amqp://%s:%s@%s:%s/", user, pass, host, port)
}

func main() {
	log.SetFormatter(&log.TextFormatter{
		ForceColors: true,
	})

	graph := graphservice.NewGraphService(getRabbitAddr())
	graph.Start(32)
	defer graph.Destroy()

	serve(graph)
}
