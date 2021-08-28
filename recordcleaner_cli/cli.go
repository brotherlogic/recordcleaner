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
	ctx, cancel := utils.ManualContext("recordcleaner-cli", time.Hour)
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
		fmt.Printf("[%v] %v\n", rec.GetRecord().GetRelease().GetInstanceId(), rec.GetRecord().GetRelease().GetTitle())
	case "refresh":
		res, err := lclient.GetClean(ctx, &pb.GetCleanRequest{IncludeSeen: true})
		if err != nil {
			log.Fatalf("Error on Get Clean: %v (%v)", err, res)
		}
		for _, id := range res.GetSeen() {
			res, err := client.ClientUpdate(ctx, &pbrc.ClientUpdateRequest{InstanceId: int32(id)})
			fmt.Printf("Refresh %v and %v\n", res, err)
		}
	case "examine":
		res, err := lclient.GetClean(ctx, &pb.GetCleanRequest{IncludeSeen: true})
		if err != nil {
			log.Fatalf("Error on Get Clean: %v (%v)", err, res)
		}
		for _, id := range res.GetSeen() {
			conn2, err2 := utils.LFDialServer(ctx, "recordcollection")
			if err2 != nil {
				log.Fatalf("%v", err)
			}
			rclient := pbrc.NewRecordCollectionServiceClient(conn2)
			rec, err := rclient.GetRecord(ctx, &pbrc.GetRecordRequest{InstanceId: id})
			if err != nil {
				log.Fatalf("%v", err)
			}
			if rec.GetRecord().GetMetadata().GetLastCleanDate() == 0 {
				log.Printf("%v", id)
			}
		}
	case "water":
		res, err := lclient.Service(ctx, &pb.ServiceRequest{Water: true})
		fmt.Printf("Watered: %v and %v\n", res, err)
	case "filter":
		res, err := lclient.Service(ctx, &pb.ServiceRequest{Fileter: true})
		fmt.Printf("Filtered: %v and %v\n", res, err)
	}
}
