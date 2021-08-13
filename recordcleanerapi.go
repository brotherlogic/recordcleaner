package main

import (
	"fmt"

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

	s.Log(fmt.Sprintf("%v", config))

	if ld, ok := config.GetLastCleanTime()[in.GetInstanceId()]; ok {
		rec, err := s.getRecord(ctx, in.GetInstanceId())
		if err != nil {
			return nil, err
		}

		if rec.GetMetadata().GetLastCleanDate() != ld {
			err := s.newClean(ctx, rec)
			if err != nil {
				return nil, err
			}
		}
	} else {
		rec, err := s.getRecord(ctx, in.GetInstanceId())
		if err != nil {
			return nil, err
		}

		err = s.newClean(ctx, rec)
		if err != nil {
			return nil, err
		}
	}

	return &rcpb.ClientUpdateResponse{}, nil
}

func (s *Server) GetClean(ctx context.Context, _ *pb.GetCleanRequest) (*pb.GetCleanResponse, error) {

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

	if len(ids.GetInstanceIds()) == 0 {
		return nil, status.Errorf(codes.ResourceExhausted, "Nothing to clean")
	}

	return &pb.GetCleanResponse{InstanceId: ids.GetInstanceIds()[0]}, nil

}
