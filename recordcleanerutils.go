package main

import (
	"time"

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
	today := 0
	for _, date := range config.GetLastCleanTime() {
		if date > 0 {
			done++
		}
		if time.Unix(date, 0).YearDay() == time.Now().YearDay() && time.Unix(date, 0).Year() == time.Now().Year() {
			today++
		}
	}

	tracked.Set(float64(total))
	cleaned.Set(float64(done))
	cleanedToday.Set(float64(today))

	cleanedLastSeven := 0
	for _, date := range config.GetLastCleanTime() {
		if time.Since(time.Unix(date, 0)) < time.Hour*24*7 {
			cleanedLastSeven++
		}
	}

	cleanedPerDay.Set(float64(cleanedLastSeven) / 7.0)
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
	config.DayCount++
	config.GetLastCleanTime()[rec.GetRelease().GetInstanceId()] = rec.GetMetadata().GetLastCleanDate()

	s.metrics(config)

	return config, s.saveConfig(ctx, config)
}
