package main

import (
	"fmt"
	"sort"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/brotherlogic/recordcleaner/proto"
	rcpb "github.com/brotherlogic/recordcollection/proto"
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

		if rec.GetMetadata().GetLastCleanDate() != ld {
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

func (s *Server) GetClean(ctx context.Context, _ *pb.GetCleanRequest) (*pb.GetCleanResponse, error) {

	config, err := s.loadConfig(ctx)
	if err != nil {
		return nil, err
	}

	if int32(time.Now().YearDay()) == config.GetDayOfYear() {
		if config.GetDayCount() > 10 {
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

	if len(ids.GetInstanceIds()) == 0 {
		return nil, status.Errorf(codes.ResourceExhausted, "Nothing to clean")
	}

	return &pb.GetCleanResponse{InstanceId: ids.GetInstanceIds()[0], Seen: sids}, nil

}
