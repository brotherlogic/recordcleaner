package main

import (
	pb "github.com/brotherlogic/recordcleaner/proto"
	rcpb "github.com/brotherlogic/recordcollection/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) metrics(config *pb.Config) {
	if config.GetLastCleanTime() == nil {
		return
	}

	total := len(config.GetLastCleanTime())
	done := 0
	for _, val := range config.GetLastCleanTime() {
		if val > 0 {
			done++
		}
	}

	tracked.Set(float64(total))
	cleaned.Set(float64(done))
	cleanedToday.Set(float64(config.GetDayCount()))
}

func (s *Server) newClean(ctx context.Context, rec *rcpb.Record) (*pb.Config, error) {
	if (rec.GetMetadata().GetFiledUnder() != rcpb.ReleaseMetadata_FILE_UNKNOWN &&
		rec.GetMetadata().GetFiledUnder() != rcpb.ReleaseMetadata_FILE_12_INCH &&
		rec.GetMetadata().GetFiledUnder() != rcpb.ReleaseMetadata_FILE_7_INCH) ||
		rec.GetMetadata().GetCategory() == rcpb.ReleaseMetadata_SOLD_ARCHIVE {
		return nil, status.Errorf(codes.InvalidArgument, "not processable")
	}

	//Run this under a lock
	key, err := s.RunLockingElection(ctx, "recordcleaner")
	if err != nil {
		return nil, err
	}

	defer s.ReleaseLockingElection(ctx, "recordcleaner", key)

	config, err := s.loadConfig(ctx)
	if err != nil {
		return nil, err
	}

	config.CurrentCount++
	config.GetLastCleanTime()[rec.GetRelease().GetInstanceId()] = rec.GetMetadata().GetLastCleanDate()

	s.metrics(config)

	return config, s.saveConfig(ctx, config)
}
