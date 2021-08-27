package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/brotherlogic/goserver"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	dspb "github.com/brotherlogic/dstore/proto"
	pbg "github.com/brotherlogic/goserver/proto"
	"github.com/brotherlogic/goserver/utils"
	pb "github.com/brotherlogic/recordcleaner/proto"
	rcpb "github.com/brotherlogic/recordcollection/proto"
	google_protobuf "github.com/golang/protobuf/ptypes/any"
)

const (
	CONFIG_KEY = "github.com/brotherlogic/recordcleaner/config"
)

var (
	tracked = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "recordcleaner_tracked",
		Help: "The size of the print queue",
	})
	cleaned = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "recordcleaner_cleaned",
		Help: "The size of the print queue",
	})
	cleanedToday = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "recordcleaner_today",
		Help: "The size of the print queue",
	})
)

//Server main server type
type Server struct {
	*goserver.GoServer
}

func (s *Server) loadConfig(ctx context.Context) (*pb.Config, error) {
	conn, err := s.FDialServer(ctx, "dstore")
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	client := dspb.NewDStoreServiceClient(conn)
	res, err := client.Read(ctx, &dspb.ReadRequest{Key: CONFIG_KEY})
	if err != nil {
		if status.Convert(err).Code() == codes.NotFound {
			return &pb.Config{LastCleanTime: make(map[int32]int64)}, nil
		}

		return nil, err

	}

	if res.GetConsensus() < 0.5 {
		return nil, fmt.Errorf("could not get read consensus (%v)", res.GetConsensus())
	}

	config := &pb.Config{}
	err = proto.Unmarshal(res.GetValue().GetValue(), config)
	if err != nil {
		return nil, err
	}
	if config.GetLastCleanTime() == nil {
		config.LastCleanTime = make(map[int32]int64)
	}

	s.metrics(config)

	return config, nil
}

func (s *Server) saveConfig(ctx context.Context, config *pb.Config) error {
	conn, err := s.FDialServer(ctx, "dstore")
	if err != nil {
		return err
	}
	defer conn.Close()

	data, err := proto.Marshal(config)
	if err != nil {
		return err
	}

	client := dspb.NewDStoreServiceClient(conn)
	res, err := client.Write(ctx, &dspb.WriteRequest{Key: CONFIG_KEY, Value: &google_protobuf.Any{Value: data}})
	if err != nil {
		return err
	}

	if res.GetConsensus() < 0.5 {
		return fmt.Errorf("could not get write consensus (%v)", res.GetConsensus())
	}

	s.metrics(config)

	return nil
}

// Init builds the server
func Init() *Server {
	s := &Server{
		GoServer: &goserver.GoServer{},
	}

	return s
}

func (s *Server) getRecord(ctx context.Context, iid int32) (*rcpb.Record, error) {
	conn, err := s.FDialServer(ctx, "recordcollection")
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	client := rcpb.NewRecordCollectionServiceClient(conn)
	r, err := client.GetRecord(ctx, &rcpb.GetRecordRequest{InstanceId: iid})
	if err != nil {
		return nil, err
	}
	return r.GetRecord(), nil
}

// DoRegister does RPC registration
func (s *Server) DoRegister(server *grpc.Server) {
	rcpb.RegisterClientUpdateServiceServer(server, s)
	pb.RegisterRecordCleanerServiceServer(server, s)
}

// ReportHealth alerts if we're not healthy
func (s *Server) ReportHealth() bool {
	return true
}

// Shutdown the server
func (s *Server) Shutdown(ctx context.Context) error {
	return nil
}

// Mote promotes/demotes this server
func (s *Server) Mote(ctx context.Context, master bool) error {
	return nil
}

// GetState gets the state of the server
func (s *Server) GetState() []*pbg.State {
	return []*pbg.State{}
}

func main() {
	var quiet = flag.Bool("quiet", false, "Show all output")
	flag.Parse()

	//Turn off logging
	if *quiet {
		log.SetFlags(0)
		log.SetOutput(ioutil.Discard)
	}
	server := Init()
	server.PrepServer()
	server.Register = server

	err := server.RegisterServerV2("recordcleaner", false, true)
	if err != nil {
		return
	}

	// Fill the mtrics out before starting
	ctx, cancel := utils.ManualContext("recordcleaner-startup", time.Minute)
	server.loadConfig(ctx)
	cancel()

	fmt.Printf("%v", server.Serve())
}
