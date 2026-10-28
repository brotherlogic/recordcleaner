package main

import (
	"net"
	"testing"

	dspb "github.com/brotherlogic/dstore/proto"
	gdpb "github.com/brotherlogic/godiscogs/proto"
	pb "github.com/brotherlogic/recordcleaner/proto"
	rcpb "github.com/brotherlogic/recordcollection/proto"
	ropb "github.com/brotherlogic/recordsorganiser/proto"
	google_protobuf "github.com/golang/protobuf/ptypes/any"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type testOrganiser struct {
	widths map[string]float32
	err    error
}

func (t *testOrganiser) getLocationWidth(ctx context.Context, name string) (float32, error) {
	if t.err != nil {
		return 0, t.err
	}
	w, ok := t.widths[name]
	if !ok {
		return 0, status.Errorf(codes.NotFound, "location %v not found", name)
	}
	return w, nil
}

type testRecordCollection struct {
	folders map[int32][]int64
	records map[int64]*rcpb.Record
	updates []*rcpb.UpdateRecordRequest
	err     error
}

func (t *testRecordCollection) queryRecords(ctx context.Context, folderID int32) ([]int64, error) {
	if t.err != nil {
		return nil, t.err
	}
	return t.folders[folderID], nil
}

func (t *testRecordCollection) getRecord(ctx context.Context, iid int64) (*rcpb.Record, error) {
	if t.err != nil {
		return nil, t.err
	}
	rec, ok := t.records[iid]
	if !ok {
		return nil, status.Errorf(codes.OutOfRange, "record %v not found", iid)
	}
	return rec, nil
}

func (t *testRecordCollection) updateRecord(ctx context.Context, req *rcpb.UpdateRecordRequest) error {
	if t.err != nil {
		return t.err
	}
	t.updates = append(t.updates, req)
	return nil
}

type testDstore struct {
	data      map[string]*google_protobuf.Any
	consensus float32
	err       error
}

func (t *testDstore) read(ctx context.Context, req *dspb.ReadRequest) (*dspb.ReadResponse, error) {
	if t.err != nil {
		return nil, t.err
	}
	val, ok := t.data[req.GetKey()]
	if !ok {
		return nil, status.Errorf(codes.InvalidArgument, "key %v not found", req.GetKey())
	}
	c := t.consensus
	if c == 0 {
		c = 1.0
	}
	return &dspb.ReadResponse{
		Value:     val,
		Consensus: c,
	}, nil
}

func (t *testDstore) write(ctx context.Context, req *dspb.WriteRequest) (*dspb.WriteResponse, error) {
	if t.err != nil {
		return nil, t.err
	}
	if t.data == nil {
		t.data = make(map[string]*google_protobuf.Any)
	}
	t.data[req.GetKey()] = req.GetValue()
	c := t.consensus
	if c == 0 {
		c = 1.0
	}
	return &dspb.WriteResponse{
		Consensus: c,
	}, nil
}

func getTestServer() *Server {
	s := Init()
	s.SkipLog = true
	s.org = &testOrganiser{
		widths: map[string]float32{
			"12 Inch Listening Pile": 100,
			"12 Inch Holding Pile":   100,
		},
	}
	s.rc = &testRecordCollection{
		folders: make(map[int32][]int64),
		records: make(map[int64]*rcpb.Record),
	}
	s.dstore = &testDstore{
		data:      make(map[string]*google_protobuf.Any),
		consensus: 1.0,
	}
	return s
}

func TestInit(t *testing.T) {
	s := Init()
	if s == nil {
		t.Fatalf("Init() returned a nil server")
	}

	if s.lastUpdate == nil {
		t.Errorf("Init() did not initialize lastUpdate map")
	}
	if s.org == nil {
		t.Errorf("Init() did not initialize org bridge")
	}
	if s.rc == nil {
		t.Errorf("Init() did not initialize rc bridge")
	}
	if s.dstore == nil {
		t.Errorf("Init() did not initialize dstore bridge")
	}
}

func TestLoadAndSaveConfig(t *testing.T) {
	s := getTestServer()
	ctx := context.Background()

	// Initial load when key does not exist should return an empty initialized Config.
	cfg, err := s.loadConfig(ctx)
	if err != nil {
		t.Fatalf("loadConfig() unexpected error on missing key: %v", err)
	}
	if cfg.GetLastCleanTime() == nil {
		t.Errorf("loadConfig() did not initialize LastCleanTime map")
	}

	// Save modified config and reload it.
	cfg.CurrentBoxPick = 42
	cfg.OutOfSleeves = true
	cfg.LastCleanTime[100] = 123456789
	if err := s.saveConfig(ctx, cfg); err != nil {
		t.Fatalf("saveConfig() unexpected error: %v", err)
	}

	loaded, err := s.loadConfig(ctx)
	if err != nil {
		t.Fatalf("loadConfig() unexpected error after save: %v", err)
	}
	if loaded.GetCurrentBoxPick() != 42 {
		t.Errorf("expected CurrentBoxPick = 42, got %v", loaded.GetCurrentBoxPick())
	}
	if !loaded.GetOutOfSleeves() {
		t.Errorf("expected OutOfSleeves = true, got false")
	}
	if loaded.GetLastCleanTime()[100] != 123456789 {
		t.Errorf("expected LastCleanTime[100] = 123456789, got %v", loaded.GetLastCleanTime()[100])
	}
}

func TestRecordCollectionBridgeOperations(t *testing.T) {
	s := getTestServer()
	ctx := context.Background()

	rcFake := s.rc.(*testRecordCollection)
	rcFake.records[101] = &rcpb.Record{
		Release: &gdpb.Release{InstanceId: 101},
		Metadata: &rcpb.ReleaseMetadata{
			FiledUnder:    rcpb.ReleaseMetadata_FILE_12_INCH,
			Category:      rcpb.ReleaseMetadata_UNLISTENED,
			LastCleanDate: 0,
			GoalFolder:    242017,
		},
	}
	rcFake.records[102] = &rcpb.Record{
		Release: &gdpb.Release{InstanceId: 102},
		Metadata: &rcpb.ReleaseMetadata{
			FiledUnder:    rcpb.ReleaseMetadata_FILE_12_INCH,
			Category:      rcpb.ReleaseMetadata_IN_COLLECTION,
			LastCleanDate: 1000,
			GoalFolder:    242017,
		},
	}
	rcFake.folders[TOGO_FOLDER] = []int64{101, 102}

	rec, err := s.getRecord(ctx, 101)
	if err != nil {
		t.Fatalf("getRecord(101) unexpected error: %v", err)
	}
	if rec.GetRelease().GetInstanceId() != 101 {
		t.Errorf("expected instance ID 101, got %v", rec.GetRelease().GetInstanceId())
	}

	s.pingRecord(ctx, 101)
	if len(rcFake.updates) != 1 {
		t.Fatalf("expected 1 updateRecord call from pingRecord, got %d", len(rcFake.updates))
	}
	if rcFake.updates[0].GetUpdate().GetRelease().GetInstanceId() != 101 || !rcFake.updates[0].GetUpdate().GetMetadata().GetNeedsGramUpdate() {
		t.Errorf("unexpected updateRecord payload: %+v", rcFake.updates[0])
	}
	if s.lastUpdate[101] == 0 {
		t.Errorf("expected lastUpdate[101] to be set after pingRecord")
	}

	if err := s.triggerMetrics(ctx); err != nil {
		t.Fatalf("triggerMetrics() unexpected error: %v", err)
	}

	rcFake.records[101].Metadata.DateArrived = 100
	resp, err := s.GetCleanInternal(ctx, &pb.GetCleanRequest{})
	if err != nil {
		t.Fatalf("GetCleanInternal() unexpected error: %v", err)
	}
	if resp.GetInstanceId() != 101 {
		t.Errorf("expected GetCleanInternal() to return 101, got %v", resp.GetInstanceId())
	}
}

type fakeOrganiserServer struct {
	ropb.UnimplementedOrganiserServiceServer
	locations map[string]*ropb.Location
}

func (f *fakeOrganiserServer) GetOrganisation(ctx context.Context, req *ropb.GetOrganisationRequest) (*ropb.GetOrganisationResponse, error) {
	var locs []*ropb.Location
	for _, l := range req.GetLocations() {
		if found, ok := f.locations[l.GetName()]; ok {
			locs = append(locs, found)
		}
	}
	return &ropb.GetOrganisationResponse{Locations: locs}, nil
}

func TestProdOrganiserGetLocationWidth(t *testing.T) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	ropb.RegisterOrganiserServiceServer(grpcServer, &fakeOrganiserServer{
		locations: map[string]*ropb.Location{
			"12 Inch Listening Pile": {
				Name: "12 Inch Listening Pile",
				ReleasesLocation: []*ropb.ReleasePlacement{
					{InstanceId: 1, DeterminedWidth: 120.5},
					{InstanceId: 2, DeterminedWidth: 230.25},
				},
			},
		},
	})
	go grpcServer.Serve(lis)
	defer grpcServer.Stop()

	org := &prodOrganiser{
		dial: func(ctx context.Context, server string) (*grpc.ClientConn, error) {
			if server != "recordsorganiser" {
				return nil, status.Errorf(codes.InvalidArgument, "unexpected server %v", server)
			}
			return grpc.NewClient(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
		},
	}

	ctx := context.Background()
	width, err := org.getLocationWidth(ctx, "12 Inch Listening Pile")
	if err != nil {
		t.Fatalf("getLocationWidth() unexpected error: %v", err)
	}
	if width != 350.75 {
		t.Errorf("expected width 350.75, got %v", width)
	}

	_, err = org.getLocationWidth(ctx, "Missing Pile")
	if status.Convert(err).Code() != codes.NotFound {
		t.Errorf("expected NotFound error for missing location, got %v", err)
	}
}

