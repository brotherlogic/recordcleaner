package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/brotherlogic/goserver/utils"
	"google.golang.org/grpc/resolver"

	pb "github.com/brotherlogic/recordcleaner/proto"
	pbrc "github.com/brotherlogic/recordcollection/proto"
)

func init() {
	resolver.Register(&utils.DiscoveryClientResolverBuilder{})
}

func main() {
	ctx, cancel := utils.ManualContext("recordcleaner-cli", time.Second*10)
	defer cancel()

	conn, err := utils.LFDialServer(ctx, "recordcleaner")
	if err != nil {
		log.Fatalf("Unable to dial: %v", err)
	}
	defer conn.Close()

	client := pbrc.NewClientUpdateServiceClient(conn)
	lclient := pb.NewRecordCleanerServiceClient(conn)

	switch os.Args[1] {
	case "update":
		updateFlags := flag.NewFlagSet("Update", flag.ExitOnError)
		var id = updateFlags.Int("id", -1, "Id of the record to add")

		if err := updateFlags.Parse(os.Args[2:]); err == nil {
			if *id > 0 {
				res, err := client.ClientUpdate(ctx, &pbrc.ClientUpdateRequest{InstanceId: int32(*id)})
				if err != nil {
					log.Fatalf("Error on Add Record: %v", err)
				}
				fmt.Printf("%v and %v", res, err)
			}
		}
	case "get":
		res, err := lclient.GetClean(ctx, &pb.GetCleanRequest{})
		if err != nil {
			log.Fatalf("Error on Get Clean: %v", err)
		}
		conn2, err2 := utils.LFDialServer(ctx, "recordcollection")
		if err2 != nil {
			log.Fatalf("Cannot find rc: %v", err2)
		}
		rcclient := pbrc.NewRecordCollectionServiceClient(conn2)
		rec, err := rcclient.GetRecord(ctx, &pbrc.GetRecordRequest{InstanceId: res.GetInstanceId()})
		if err != nil {
			log.Fatalf("Cannot get record: %v", err)
		}
		fmt.Printf("%v\n", rec.GetRecord().GetRelease().GetTitle())
	}
}
