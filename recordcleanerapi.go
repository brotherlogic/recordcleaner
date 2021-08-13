package main

import (
	"fmt"

	"golang.org/x/net/context"

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
