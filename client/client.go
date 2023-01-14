package recordcleaner_client

import (
	"context"

	pbgs "github.com/brotherlogic/goserver"
	pb "github.com/brotherlogic/recordcleaner/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type RecordCleanerClient struct {
	Gs        *pbgs.GoServer
	ErrorCode codes.Code
	Test      bool
}

func (c *RecordCleanerClient) GetClean(ctx context.Context, in *pb.GetCleanRequest) (*pb.GetCleanResponse, error) {
	if c.Test {
		if c.ErrorCode != codes.OK {
			return nil, status.Errorf(c.ErrorCode, "Forced Error")
		}
		return &pb.GetCleanResponse{InstanceId: 1234}, nil
	}
	conn, err := c.Gs.FDialServer(ctx, "recordcleaner")
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	client := pb.NewRecordCleanerServiceClient(conn)
	return client.GetClean(ctx, in)
}
