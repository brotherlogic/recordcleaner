package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/brotherlogic/goserver"
	"github.com/brotherlogic/goserver/utils"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	gdpb "github.com/brotherlogic/godiscogs"
	pbg "github.com/brotherlogic/goserver/proto"
	rcpb "github.com/brotherlogic/recordcollection/proto"
	pb "github.com/brotherlogic/recordvalidator/proto"
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
)

//Server main server type
type Server struct {
	*goserver.GoServer
}

// Init builds the server
func Init() *Server {

	s := &Server{
		GoServer: &goserver.GoServer{},
	}

	return s
}






func (s *Server) getRecord(ctx context.Context, iid int32) (*rcpb.Record, error) {
	if s.failLoadAll {
		return nil, fmt.Errorf("Built to fail")
	}
	if s.test {
		return &rcpb.Record{}, nil
	}
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
	return []*pbg.State{
		&pbg.State{Key: "magic", Text: fmt.Sprintf("%v", found)},
	}
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

	//Do a load to prepopulate metrics
	ctx, cancel := utils.ManualContext("rvsu", time.Minute)
	if _, err := server.load(ctx); err != nil {
		server.Log(fmt.Sprintf("Unable to load: %v", err))
		time.Sleep(time.Second * 5)
		return
	}
	cancel()

	fmt.Printf("%v", server.Serve())
}
