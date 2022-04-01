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
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

var getEnv = utils.GetEnv
var failOnError = utils.FailOnError

type pathfinderServer struct {
	pb.UnimplementedPathfinderServer

	g *graphservice.GraphService
}

func (s *pathfinderServer) bfs(ctx context.Context, URL1 string, URL2 string, reqId string) ([]string, error) {
	sublogger := log.With().Str("id", reqId).Logger()

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
					sublogger.Info().Str("page", last).Msg("Checking neighbors")
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

			log.Info().
				Int("breadth", breadth).
				Int("N", N).
				Msg("Started fetching for new breadth level")

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
				sublogger.Error().Err(err).Msg("")
				return
			}
		}
	}()

	select {
	case <-ctx.Done():
		sublogger.Info().Msg("BFS timed out")
		return nil, errors.New("Request took too long")
	case path := <-doneChan:
		return path, nil
	}
}

func validateWikiUrl(urlStr string) error {
	u, err := url.Parse(urlStr)
	if err != nil {
		return err
	}

	parts := strings.Split(u.Hostname(), ".")
	if len(parts) < 2 || parts[len(parts)-2] != "wikipedia" || parts[len(parts)-1] != "org" {
		err = fmt.Errorf("Expected wikipedia url, but received %s", urlStr)
		return err
	}

	return nil
}

func (s *pathfinderServer) Find(ctx context.Context, in *pb.FindRequest) (*pb.FindResponse, error) {
	reqId := uuid.NewString()

	sublogger := log.With().
		Str("URL1", in.URL1).
		Str("URL2", in.URL2).
		Str("id", reqId).
		Logger()
	sublogger.Info().Msg("Processing request")
	if err := validateWikiUrl(in.URL1); err != nil {
		sublogger.Info().AnErr("err", err).Msg("URL1 validation failed")
		return nil, err
	}
	if err := validateWikiUrl(in.URL2); err != nil {
		sublogger.Info().AnErr("err", err).Msg("URL2 validation failed")
		return nil, err
	}
	if in.URL1 == in.URL2 {
		sublogger.Info().Msg("Finished processing request")
		return &pb.FindResponse{
			Length: 0,
			Path:   []string{in.URL1},
		}, nil
	}

	ctx, cancel := context.WithTimeout(ctx, time.Duration(120)*time.Second)
	defer cancel()
	path, err := s.bfs(ctx, in.URL1, in.URL2, reqId)
	if err != nil {
		return nil, err
	}

	sublogger.Info().Msg("Finished processing request")
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
		log.Fatal().AnErr("err", err).Msg("Failed to listen")
	}
	log.Info().Str("addr", lis.Addr().String()).Msg("Listening")
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
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	graph := graphservice.NewGraphService(getRabbitAddr())
	graph.Start(32)
	defer graph.Destroy()

	serve(graph)
}
