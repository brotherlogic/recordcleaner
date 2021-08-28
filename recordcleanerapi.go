package main

import (
	"fmt"
	"math/rand"
	"sort"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/brotherlogic/recordcleaner/proto"
	rcpb "github.com/brotherlogic/recordcollection/proto"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	water = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "recordcleaner_water",
		Help: "The size of the print queue",
	})
	filter = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "recordcleaner_filter",
		Help: "The size of the print queue",
	})
)

//ClientUpdate forces a move
func (s *Server) ClientUpdate(ctx context.Context, in *rcpb.ClientUpdateRequest) (*rcpb.ClientUpdateResponse, error) {
	config, err := s.loadConfig(ctx)
	if err != nil {
		return nil, err
	}

	if config.GetLastCleanTime() == nil {
		config.LastCleanTime = make(map[int32]int64)
	}

	if ld, ok := config.GetLastCleanTime()[in.GetInstanceId()]; ok {
		rec, err := s.getRecord(ctx, in.GetInstanceId())
		if err != nil {
			if status.Convert(err).Code() == codes.OutOfRange {
				delete(config.LastCleanTime, in.GetInstanceId())
				return &rcpb.ClientUpdateResponse{}, s.saveConfig(ctx, config)
			}

			return nil, err
		}

		if config.GetCurrentBoxPick() == in.GetInstanceId() {
			config.CurrentBoxPick = 0
		}

		if ld == 0 {
			s.Log(fmt.Sprintf("UNCLEAN %v", in.GetInstanceId()))
		}

		if (rec.GetMetadata().GetFiledUnder() != rcpb.ReleaseMetadata_FILE_UNKNOWN &&
			rec.GetMetadata().GetFiledUnder() != rcpb.ReleaseMetadata_FILE_12_INCH &&
			rec.GetMetadata().GetFiledUnder() != rcpb.ReleaseMetadata_FILE_7_INCH) ||
			rec.GetMetadata().GetCategory() == rcpb.ReleaseMetadata_SOLD_ARCHIVE {
			delete(config.LastCleanTime, in.GetInstanceId())

			err = s.saveConfig(ctx, config)
			if err != nil {
				return nil, err
			}

		} else if rec.GetMetadata().GetLastCleanDate() != ld {
			config, err := s.newClean(ctx, rec)
			if err != nil {
				return nil, err
			}

			if time.Now().YearDay() == int(config.GetDayOfYear()) {
				config.DayCount++
			} else {
				config.DayCount = 1
				config.DayOfYear = int32(time.Now().YearDay())
			}

			s.Log(fmt.Sprintf("Day clean %v and %v and %v from %v (since %v and %v) => %v", config.DayCount, config.DayOfYear, time.Now().YearDay(), in.GetInstanceId(), rec.GetMetadata().GetLastCleanDate(), ld, config.GetLastCleanTime()[in.GetInstanceId()]))

			err = s.saveConfig(ctx, config)
			if err != nil {
				return nil, err
			}
		}
	} else {
		rec, err := s.getRecord(ctx, in.GetInstanceId())
		if err != nil {
			return nil, err
		}

		_, err = s.newClean(ctx, rec)
		if err != nil {
			return nil, err
		}
	}

	return &rcpb.ClientUpdateResponse{}, nil
}

func (s *Server) Service(ctx context.Context, req *pb.ServiceRequest) (*pb.ServiceResponse, error) {
	config, err := s.loadConfig(ctx)
	if err != nil {
		return nil, err
	}

	if req.GetFileter() {
		config.LastFilter = time.Now().Unix()
	}
	if req.GetWater() {
		config.LastWater = time.Now().Unix()
	}

	return &pb.ServiceResponse{}, s.saveConfig(ctx, config)
}

func (s *Server) GetClean(ctx context.Context, req *pb.GetCleanRequest) (*pb.GetCleanResponse, error) {

	config, err := s.loadConfig(ctx)
	if err != nil {
		return nil, err
	}

	waterCount := 0
	filterCount := 0
	for _, date := range config.GetLastCleanTime() {
		if date > int64(config.GetLastWater()) {
			waterCount++
		}
		if date > int64(config.GetLastFilter()) {
			filterCount++
		}
	}
	water.Set(float64(waterCount))
	filter.Set(float64(filterCount))

	if waterCount >= 30 {
		return nil, status.Errorf(codes.FailedPrecondition, "You need to change the water, it was last done on %v", time.Unix(config.GetLastWater(), 0))
	}

	if filterCount >= 50 {
		return nil, status.Errorf(codes.FailedPrecondition, "You need to change the filter, it was last done on %v", time.Unix(config.GetLastFilter(), 0))
	}

	if int32(time.Now().YearDay()) == config.GetDayOfYear() {
		if config.GetDayCount() >= 10 {
			return nil, status.Errorf(codes.FailedPrecondition, "you've cleaned %v records today, that's plenty", config.GetDayCount())
		}
	}

	conn, err := s.FDialServer(ctx, "recordcollection")
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	client := rcpb.NewRecordCollectionServiceClient(conn)
	ids, err := client.QueryRecords(ctx, &rcpb.QueryRecordsRequest{Query: &rcpb.QueryRecordsRequest_FolderId{int32(3386035)}})
	if err != nil {
		return nil, err
	}

	sort.SliceStable(ids.InstanceIds, func(i, j int) bool {
		return ids.InstanceIds[i] < ids.InstanceIds[j]
	})

	var sids []int32
	for id, _ := range config.GetLastCleanTime() {
		sids = append(sids, id)
	}

	if !req.GetIncludeSeen() && len(ids.GetInstanceIds()) == 0 {
		if config.GetCurrentBoxPick() == 0 {
			ids, err := client.QueryRecords(ctx, &rcpb.QueryRecordsRequest{Query: &rcpb.QueryRecordsRequest_FolderId{int32(3282985)}})
			if err != nil {
				return nil, err
			}

			if len(ids.GetInstanceIds()) == 0 {
				return nil, status.Errorf(codes.ResourceExhausted, "Nothing to clean")
			}

			config.CurrentBoxPick = ids.GetInstanceIds()[rand.Intn(len(ids.GetInstanceIds()))]
			err = s.saveConfig(ctx, config)
			if err != nil {
				return nil, err
			}
		}

		return &pb.GetCleanResponse{InstanceId: config.CurrentBoxPick, Seen: sids}, nil
	}

	if len(ids.GetInstanceIds()) == 0 {
		return &pb.GetCleanResponse{Seen: sids}, nil
	}

	return &pb.GetCleanResponse{InstanceId: ids.GetInstanceIds()[0], Seen: sids}, nil
}
