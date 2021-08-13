package main

import (
	pb "github.com/brotherlogic/recordcleaner/proto"
	rcpb "github.com/brotherlogic/recordcollection/proto"
	"golang.org/x/net/context"
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
}

func (s *Server) newClean(ctx context.Context, rec *rcpb.Record) error {
	//Run this under a lock
	key, err := s.RunLockingElection(ctx, "recordcleaner")
	if err != nil {
		return err
	}

	defer s.ReleaseLockingElection(ctx, "recordcleaner", key)

	config, err := s.loadConfig(ctx)
	if err != nil {
		return err
	}

	config.CurrentCount++
	config.GetLastCleanTime()[rec.GetRelease().GetInstanceId()] = rec.GetMetadata().GetLastCleanDate()

	s.metrics(config)

	return s.saveConfig(ctx, config)
}
