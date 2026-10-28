package main

import (
	"fmt"
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
	gdpb "github.com/brotherlogic/godiscogs/proto"
	pbg "github.com/brotherlogic/goserver/proto"
	"github.com/brotherlogic/goserver/utils"
	pb "github.com/brotherlogic/recordcleaner/proto"
	rcpb "github.com/brotherlogic/recordcollection/proto"
	ropb "github.com/brotherlogic/recordsorganiser/proto"
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
	cleanedTodayCat = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "recordcleaner_today_cat",
		Help: "The size of the print queue",
	}, []string{"category"})
)

type organiserBridge interface {
	getLocationWidth(ctx context.Context, name string) (float32, error)
}

type recordCollectionBridge interface {
	queryRecords(ctx context.Context, folderID int32) ([]int64, error)
	getRecord(ctx context.Context, iid int64) (*rcpb.Record, error)
	updateRecord(ctx context.Context, req *rcpb.UpdateRecordRequest) error
}

type dstoreBridge interface {
	read(ctx context.Context, req *dspb.ReadRequest) (*dspb.ReadResponse, error)
	write(ctx context.Context, req *dspb.WriteRequest) (*dspb.WriteResponse, error)
}

type prodOrganiser struct {
	dial func(ctx context.Context, server string) (*grpc.ClientConn, error)
}

func (p *prodOrganiser) getLocationWidth(ctx context.Context, name string) (float32, error) {
	conn, err := p.dial(ctx, "recordsorganiser")
	if err != nil {
		return 0, err
	}
	defer conn.Close()

	client := ropb.NewOrganiserServiceClient(conn)
	resp, err := client.GetOrganisation(ctx, &ropb.GetOrganisationRequest{
		Locations: []*ropb.Location{{Name: name}},
	})
	if err != nil {
		return 0, err
	}
	if len(resp.GetLocations()) == 0 {
		return 0, status.Errorf(codes.NotFound, "location %v not found", name)
	}

	var total float32
	for _, placement := range resp.GetLocations()[0].GetReleasesLocation() {
		total += placement.GetDeterminedWidth()
	}
	return total, nil
}

type prodRecordCollection struct {
	dial func(ctx context.Context, server string) (*grpc.ClientConn, error)
}

func (p *prodRecordCollection) queryRecords(ctx context.Context, folderID int32) ([]int64, error) {
	conn, err := p.dial(ctx, "recordcollection")
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	client := rcpb.NewRecordCollectionServiceClient(conn)
	resp, err := client.QueryRecords(ctx, &rcpb.QueryRecordsRequest{
		Query: &rcpb.QueryRecordsRequest_FolderId{FolderId: folderID},
	})
	if err != nil {
		return nil, err
	}
	return resp.GetInstanceIds(), nil
}

func (p *prodRecordCollection) getRecord(ctx context.Context, iid int64) (*rcpb.Record, error) {
	conn, err := p.dial(ctx, "recordcollection")
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	client := rcpb.NewRecordCollectionServiceClient(conn)
	resp, err := client.GetRecord(ctx, &rcpb.GetRecordRequest{InstanceId: iid})
	if err != nil {
		return nil, err
	}
	return resp.GetRecord(), nil
}

func (p *prodRecordCollection) updateRecord(ctx context.Context, req *rcpb.UpdateRecordRequest) error {
	conn, err := p.dial(ctx, "recordcollection")
	if err != nil {
		return err
	}
	defer conn.Close()

	client := rcpb.NewRecordCollectionServiceClient(conn)
	_, err = client.UpdateRecord(ctx, req)
	return err
}

type prodDstore struct {
	dial func(ctx context.Context, server string) (*grpc.ClientConn, error)
}

func (p *prodDstore) read(ctx context.Context, req *dspb.ReadRequest) (*dspb.ReadResponse, error) {
	conn, err := p.dial(ctx, "dstore")
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	client := dspb.NewDStoreServiceClient(conn)
	return client.Read(ctx, req)
}

func (p *prodDstore) write(ctx context.Context, req *dspb.WriteRequest) (*dspb.WriteResponse, error) {
	conn, err := p.dial(ctx, "dstore")
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	client := dspb.NewDStoreServiceClient(conn)
	return client.Write(ctx, req)
}

// Server main server type
type Server struct {
	*goserver.GoServer
	lastUpdate map[int64]int64
	org        organiserBridge
	rc         recordCollectionBridge
	dstore     dstoreBridge
}

func (s *Server) loadConfig(ctx context.Context) (*pb.Config, error) {
	res, err := s.dstore.read(ctx, &dspb.ReadRequest{Key: CONFIG_KEY})
	if err != nil {
		if status.Convert(err).Code() == codes.InvalidArgument {
			return &pb.Config{LastCleanTime: make(map[int64]int64)}, nil
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
		config.LastCleanTime = make(map[int64]int64)
	}
	if config.GetDayCategoryCount() == nil {
		config.DayCategoryCount = map[int64]string{}
	}

	s.metrics(ctx, config)

	return config, nil
}

func (s *Server) saveConfig(ctx context.Context, config *pb.Config) error {
	data, err := proto.Marshal(config)
	if err != nil {
		return err
	}

	res, err := s.dstore.write(ctx, &dspb.WriteRequest{Key: CONFIG_KEY, Value: &google_protobuf.Any{Value: data}})
	if err != nil {
		return err
	}

	if res.GetConsensus() < 0.5 {
		return fmt.Errorf("could not get write consensus (%v)", res.GetConsensus())
	}

	s.metrics(ctx, config)

	return nil
}

// Init builds the server
func Init() *Server {
	s := &Server{
		GoServer:   &goserver.GoServer{},
		lastUpdate: make(map[int64]int64),
	}
	s.org = &prodOrganiser{dial: s.FDialServer}
	s.rc = &prodRecordCollection{dial: s.FDialServer}
	s.dstore = &prodDstore{dial: s.FDialServer}

	return s
}

func (s *Server) getRecord(ctx context.Context, iid int64) (*rcpb.Record, error) {
	return s.rc.getRecord(ctx, iid)
}

func (s *Server) pingRecord(ctx context.Context, iid int64) {
	err := s.rc.updateRecord(ctx, &rcpb.UpdateRecordRequest{Reason: "Pulling Gram", Update: &rcpb.Record{Release: &gdpb.Release{InstanceId: iid}, Metadata: &rcpb.ReleaseMetadata{NeedsGramUpdate: true}}})
	if err == nil {
		s.lastUpdate[iid] = time.Now().Unix()
	}
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
	server := Init()
	server.PrepServer("recordcleaner")
	server.Register = server

	err := server.RegisterServerV2(false)
	if err != nil {
		return
	}

	// Fill the mtrics out before starting
	ctx, cancel := utils.ManualContext("recordcleaner-startup", time.Minute)
	server.loadConfig(ctx)
	server.triggerMetrics(ctx)
	cancel()

	fmt.Printf("%v", server.Serve())
}
