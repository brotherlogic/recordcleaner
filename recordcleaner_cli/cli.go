package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/brotherlogic/goserver/utils"

	pbrc "github.com/brotherlogic/recordcollection/proto"

	//Needed to pull in gzip encoding init
	_ "google.golang.org/grpc/encoding/gzip"
	"google.golang.org/grpc/resolver"
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
	}
}
