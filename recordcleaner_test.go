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

func TestGetCleanInternalPileWidthBoundary(t *testing.T) {
	s := getTestServer()
	ctx := context.Background()

	orgFake := s.org.(*testOrganiser)
	orgFake.widths["12 Inch Listening Pile"] = 350
	orgFake.widths["12 Inch Holding Pile"] = 350

	rcFake := s.rc.(*testRecordCollection)
	// Loose pile record: PRE_VALIDATE (non-UNLISTENED) 12-inch
	rcFake.records[201] = &rcpb.Record{
		Release: &gdpb.Release{InstanceId: 201},
		Metadata: &rcpb.ReleaseMetadata{
			FiledUnder:    rcpb.ReleaseMetadata_FILE_12_INCH,
			Category:      rcpb.ReleaseMetadata_PRE_VALIDATE,
			LastCleanDate: 0,
		},
	}
	rcFake.folders[3386035] = []int64{201}

	resp, err := s.GetCleanInternal(ctx, &pb.GetCleanRequest{})
	if err != nil {
		t.Fatalf("expected no error at boundary width 700, got %v", err)
	}
	if resp.GetInstanceId() != 201 {
		t.Errorf("expected loose pile record 201 at width 700, got %v", resp.GetInstanceId())
	}

	// Empty loose pile, test TOGO_FOLDER with non-UNLISTENED 12-inch record at boundary 700
	rcFake.folders[3386035] = nil
	rcFake.records[202] = &rcpb.Record{
		Release: &gdpb.Release{InstanceId: 202},
		Metadata: &rcpb.ReleaseMetadata{
			FiledUnder:    rcpb.ReleaseMetadata_FILE_12_INCH,
			Category:      rcpb.ReleaseMetadata_IN_COLLECTION,
			DateArrived:   100,
			LastCleanDate: 0,
			GoalFolder:    242017,
		},
	}
	rcFake.folders[TOGO_FOLDER] = []int64{202}

	resp, err = s.GetCleanInternal(ctx, &pb.GetCleanRequest{})
	if err != nil {
		t.Fatalf("expected no error for TOGO_FOLDER at boundary width 700, got %v", err)
	}
	if resp.GetInstanceId() != 202 {
		t.Errorf("expected TOGO_FOLDER record 202 at width 700, got %v", resp.GetInstanceId())
	}
}

func TestGetCleanInternalLoosePileFilteringOverWidth(t *testing.T) {
	s := getTestServer()
	ctx := context.Background()

	orgFake := s.org.(*testOrganiser)
	orgFake.widths["12 Inch Listening Pile"] = 400
	orgFake.widths["12 Inch Holding Pile"] = 300.1 // 700.1 > 700

	rcFake := s.rc.(*testRecordCollection)
	// 302 has higher ID (sorted first) and is PRE_VALIDATE (normally prioritized first),
	// 301 is UNLISTENED. Over width limit, 302 must be skipped and 301 returned.
	rcFake.records[301] = &rcpb.Record{
		Release: &gdpb.Release{InstanceId: 301},
		Metadata: &rcpb.ReleaseMetadata{
			FiledUnder:    rcpb.ReleaseMetadata_FILE_12_INCH,
			Category:      rcpb.ReleaseMetadata_UNLISTENED,
			LastCleanDate: 0,
		},
	}
	rcFake.records[302] = &rcpb.Record{
		Release: &gdpb.Release{InstanceId: 302},
		Metadata: &rcpb.ReleaseMetadata{
			FiledUnder:    rcpb.ReleaseMetadata_FILE_12_INCH,
			Category:      rcpb.ReleaseMetadata_PRE_VALIDATE,
			LastCleanDate: 0,
		},
	}
	rcFake.folders[3386035] = []int64{301, 302}

	resp, err := s.GetCleanInternal(ctx, &pb.GetCleanRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.GetInstanceId() != 301 {
		t.Errorf("expected UNLISTENED record 301 when over width limit, got %v", resp.GetInstanceId())
	}

	// Only non-UNLISTENED in 3386035: with OnlyEssential=true, should return ResourceExhausted
	rcFake.folders[3386035] = []int64{302}
	_, err = s.GetCleanInternal(ctx, &pb.GetCleanRequest{OnlyEssential: true})
	if status.Convert(err).Code() != codes.ResourceExhausted {
		t.Errorf("expected ResourceExhausted when OnlyEssential=true and loose pile only has non-UNLISTENED, got %v", err)
	}

	// Without OnlyEssential, should fall through to TOGO_FOLDER and pick UNLISTENED record 303
	rcFake.records[303] = &rcpb.Record{
		Release: &gdpb.Release{InstanceId: 303},
		Metadata: &rcpb.ReleaseMetadata{
			FiledUnder:    rcpb.ReleaseMetadata_FILE_12_INCH,
			Category:      rcpb.ReleaseMetadata_UNLISTENED,
			DateArrived:   100,
			LastCleanDate: 0,
			GoalFolder:    242017,
		},
	}
	rcFake.folders[TOGO_FOLDER] = []int64{303}

	resp, err = s.GetCleanInternal(ctx, &pb.GetCleanRequest{})
	if err != nil {
		t.Fatalf("unexpected error falling through to TOGO_FOLDER: %v", err)
	}
	if resp.GetInstanceId() != 303 {
		t.Errorf("expected fallthrough to TOGO_FOLDER record 303, got %v", resp.GetInstanceId())
	}
}

func TestGetCleanInternalTogoFolderFilteringOverWidth(t *testing.T) {
	s := getTestServer()
	ctx := context.Background()

	orgFake := s.org.(*testOrganiser)
	orgFake.widths["12 Inch Listening Pile"] = 500
	orgFake.widths["12 Inch Holding Pile"] = 250 // 750 > 700

	rcFake := s.rc.(*testRecordCollection)
	rcFake.records[401] = &rcpb.Record{
		Release: &gdpb.Release{InstanceId: 401},
		Metadata: &rcpb.ReleaseMetadata{
			FiledUnder:    rcpb.ReleaseMetadata_FILE_12_INCH,
			Category:      rcpb.ReleaseMetadata_IN_COLLECTION,
			DateArrived:   100,
			LastCleanDate: 0,
			GoalFolder:    242017,
		},
	}
	rcFake.records[402] = &rcpb.Record{
		Release: &gdpb.Release{InstanceId: 402},
		Metadata: &rcpb.ReleaseMetadata{
			FiledUnder:    rcpb.ReleaseMetadata_FILE_7_INCH,
			Category:      rcpb.ReleaseMetadata_IN_COLLECTION,
			DateArrived:   100,
			LastCleanDate: 0,
			GoalFolder:    242017,
		},
	}
	rcFake.folders[TOGO_FOLDER] = []int64{401, 402}

	// Non-UNLISTENED 12-inch (401) should be skipped, falling back to 7-inch (402) and leaving CurrentBoxPick == 0
	resp, err := s.GetCleanInternal(ctx, &pb.GetCleanRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.GetInstanceId() != 402 {
		t.Errorf("expected 7-inch fallback record 402, got %v", resp.GetInstanceId())
	}
	cfg, err := s.loadConfig(ctx)
	if err != nil {
		t.Fatalf("loadConfig error: %v", err)
	}
	if cfg.GetCurrentBoxPick() != 0 {
		t.Errorf("expected CurrentBoxPick to remain 0, got %v", cfg.GetCurrentBoxPick())
	}

	// If only non-UNLISTENED 12-inch records exist (no 7-inch), return ResourceExhausted
	rcFake.folders[TOGO_FOLDER] = []int64{401}
	_, err = s.GetCleanInternal(ctx, &pb.GetCleanRequest{})
	if status.Convert(err).Code() != codes.ResourceExhausted {
		t.Errorf("expected ResourceExhausted when no eligible TOGO_FOLDER records exist, got %v", err)
	}

	// If both non-UNLISTENED (401) and UNLISTENED (403) 12-inch records exist, only 403 is picked and saved to CurrentBoxPick
	rcFake.records[403] = &rcpb.Record{
		Release: &gdpb.Release{InstanceId: 403},
		Metadata: &rcpb.ReleaseMetadata{
			FiledUnder:    rcpb.ReleaseMetadata_FILE_12_INCH,
			Category:      rcpb.ReleaseMetadata_UNLISTENED,
			DateArrived:   100,
			LastCleanDate: 0,
			GoalFolder:    242017,
		},
	}
	rcFake.folders[TOGO_FOLDER] = []int64{401, 403}
	resp, err = s.GetCleanInternal(ctx, &pb.GetCleanRequest{})
	if err != nil {
		t.Fatalf("unexpected error picking UNLISTENED from TOGO_FOLDER: %v", err)
	}
	if resp.GetInstanceId() != 403 {
		t.Errorf("expected UNLISTENED TOGO_FOLDER record 403, got %v", resp.GetInstanceId())
	}
	cfg, err = s.loadConfig(ctx)
	if err != nil {
		t.Fatalf("loadConfig error: %v", err)
	}
	if cfg.GetCurrentBoxPick() != 403 {
		t.Errorf("expected CurrentBoxPick to be 403, got %v", cfg.GetCurrentBoxPick())
	}
}

func TestGetCleanInternalExistingBoxPickExemption(t *testing.T) {
	s := getTestServer()
	ctx := context.Background()

	orgFake := s.org.(*testOrganiser)
	orgFake.widths["12 Inch Listening Pile"] = 500
	orgFake.widths["12 Inch Holding Pile"] = 300 // 800 > 700

	rcFake := s.rc.(*testRecordCollection)
	rcFake.records[501] = &rcpb.Record{
		Release: &gdpb.Release{InstanceId: 501},
		Metadata: &rcpb.ReleaseMetadata{
			FiledUnder:    rcpb.ReleaseMetadata_FILE_12_INCH,
			Category:      rcpb.ReleaseMetadata_IN_COLLECTION,
			DateArrived:   100,
			LastCleanDate: 0,
			GoalFolder:    242017,
		},
	}

	cfg, err := s.loadConfig(ctx)
	if err != nil {
		t.Fatalf("loadConfig error: %v", err)
	}
	cfg.CurrentBoxPick = 501
	if err := s.saveConfig(ctx, cfg); err != nil {
		t.Fatalf("saveConfig error: %v", err)
	}

	resp, err := s.GetCleanInternal(ctx, &pb.GetCleanRequest{})
	if err != nil {
		t.Fatalf("expected existing CurrentBoxPick to be returned without error, got %v", err)
	}
	if resp.GetInstanceId() != 501 {
		t.Errorf("expected existing CurrentBoxPick 501, got %v", resp.GetInstanceId())
	}
}

func TestGetCleanInternalSevenInchLoosePileExemption(t *testing.T) {
	s := getTestServer()
	ctx := context.Background()

	orgFake := s.org.(*testOrganiser)
	orgFake.widths["12 Inch Listening Pile"] = 450
	orgFake.widths["12 Inch Holding Pile"] = 450 // 900 > 700

	rcFake := s.rc.(*testRecordCollection)
	rcFake.records[601] = &rcpb.Record{
		Release: &gdpb.Release{InstanceId: 601},
		Metadata: &rcpb.ReleaseMetadata{
			FiledUnder:    rcpb.ReleaseMetadata_FILE_7_INCH,
			Category:      rcpb.ReleaseMetadata_IN_COLLECTION,
			LastCleanDate: 0,
		},
	}
	rcFake.folders[3386035] = []int64{601}

	resp, err := s.GetCleanInternal(ctx, &pb.GetCleanRequest{})
	if err != nil {
		t.Fatalf("expected 7-inch record in loose pile to be returned, got error: %v", err)
	}
	if resp.GetInstanceId() != 601 {
		t.Errorf("expected 7-inch record 601, got %v", resp.GetInstanceId())
	}
}

func TestGetCleanInternalOutOfSleevesAndOverWidth(t *testing.T) {
	s := getTestServer()
	ctx := context.Background()

	orgFake := s.org.(*testOrganiser)
	orgFake.widths["12 Inch Listening Pile"] = 400
	orgFake.widths["12 Inch Holding Pile"] = 400 // 800 > 700

	cfg, err := s.loadConfig(ctx)
	if err != nil {
		t.Fatalf("loadConfig error: %v", err)
	}
	cfg.OutOfSleeves = true
	if err := s.saveConfig(ctx, cfg); err != nil {
		t.Fatalf("saveConfig error: %v", err)
	}

	rcFake := s.rc.(*testRecordCollection)
	// 701: UNLISTENED, LastCleanDate == 0 (invalid for OutOfSleeves)
	rcFake.records[701] = &rcpb.Record{
		Release: &gdpb.Release{InstanceId: 701},
		Metadata: &rcpb.ReleaseMetadata{
			FiledUnder:    rcpb.ReleaseMetadata_FILE_12_INCH,
			Category:      rcpb.ReleaseMetadata_UNLISTENED,
			LastCleanDate: 0,
		},
	}
	// 702: UNLISTENED, LastCleanDate == 100 (valid for both)
	rcFake.records[702] = &rcpb.Record{
		Release: &gdpb.Release{InstanceId: 702},
		Metadata: &rcpb.ReleaseMetadata{
			FiledUnder:    rcpb.ReleaseMetadata_FILE_12_INCH,
			Category:      rcpb.ReleaseMetadata_UNLISTENED,
			LastCleanDate: 100,
		},
	}
	// 703: IN_COLLECTION, LastCleanDate == 100 (higher ID sorted first, but invalid for overWidthLimit)
	rcFake.records[703] = &rcpb.Record{
		Release: &gdpb.Release{InstanceId: 703},
		Metadata: &rcpb.ReleaseMetadata{
			FiledUnder:    rcpb.ReleaseMetadata_FILE_12_INCH,
			Category:      rcpb.ReleaseMetadata_IN_COLLECTION,
			LastCleanDate: 100,
		},
	}
	rcFake.folders[3386035] = []int64{701, 702, 703}

	resp, err := s.GetCleanInternal(ctx, &pb.GetCleanRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.GetInstanceId() != 702 {
		t.Errorf("expected record 702 satisfying both OutOfSleeves and UNLISTENED, got %v", resp.GetInstanceId())
	}
}

func TestGetCleanInternalOrganiserErrorAndAutoRecovery(t *testing.T) {
	s := getTestServer()
	ctx := context.Background()

	orgFake := s.org.(*testOrganiser)
	delete(orgFake.widths, "12 Inch Holding Pile") // Missing location -> NotFound

	_, err := s.GetCleanInternal(ctx, &pb.GetCleanRequest{})
	if status.Convert(err).Code() != codes.NotFound {
		t.Errorf("expected NotFound when holding pile location is missing, got %v", err)
	}

	orgFake.err = status.Errorf(codes.Unavailable, "organiser unavailable")
	_, err = s.GetCleanInternal(ctx, &pb.GetCleanRequest{})
	if status.Convert(err).Code() != codes.Unavailable {
		t.Errorf("expected Unavailable error from organiser to propagate, got %v", err)
	}
	orgFake.err = nil

	// Over width limit (> 700) blocks non-UNLISTENED 12-inch record 801
	orgFake.widths["12 Inch Listening Pile"] = 400
	orgFake.widths["12 Inch Holding Pile"] = 350 // 750 > 700

	rcFake := s.rc.(*testRecordCollection)
	rcFake.records[801] = &rcpb.Record{
		Release: &gdpb.Release{InstanceId: 801},
		Metadata: &rcpb.ReleaseMetadata{
			FiledUnder:    rcpb.ReleaseMetadata_FILE_12_INCH,
			Category:      rcpb.ReleaseMetadata_PRE_VALIDATE,
			LastCleanDate: 0,
		},
	}
	rcFake.folders[3386035] = []int64{801}

	_, err = s.GetCleanInternal(ctx, &pb.GetCleanRequest{})
	if status.Convert(err).Code() != codes.ResourceExhausted {
		t.Errorf("expected ResourceExhausted when combinedWidth > 700, got %v", err)
	}

	// Automatic recovery when combinedWidth drops back to <= 700
	orgFake.widths["12 Inch Holding Pile"] = 250 // 650 <= 700
	resp, err := s.GetCleanInternal(ctx, &pb.GetCleanRequest{})
	if err != nil {
		t.Fatalf("expected automatic recovery when combinedWidth <= 700, got %v", err)
	}
	if resp.GetInstanceId() != 801 {
		t.Errorf("expected record 801 after recovery, got %v", resp.GetInstanceId())
	}
}


