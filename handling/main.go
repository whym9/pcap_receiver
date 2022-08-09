package main

import (
	"flag"
	"log"
	"os"
	process "pcap_receiver/handling/process"
	"pcap_receiver/handling/servers/GRPC"
	"pcap_receiver/handling/servers/HTTP"
	"pcap_receiver/handling/servers/TCP"
)

var DBMS string
var DBName string
var address string
var uploadDir string
var sender string

func main() {
	DBMS = *flag.String("DBMS", "mysql", "Choose DBMS")
	DBName = *flag.String("DBName", "myserver_dns", "Choose DB")
	address = *flag.String("address", "localhost:8080", "choose address")
	uploadDir = *flag.String("dir", "files", "choose upload directory")
	sender = *flag.String("sender", "GRPC", "choose sender")
	flag.Parse()
	err := os.MkdirAll(uploadDir, os.ModePerm)

	if err != nil {
		log.Fatalf("couldn't create path, %v", err)
		return
	}
	DB := process.Saver{}
	DB.Create(DBMS, DBName)

	switch sender {
	case "TCP":
		TCP.HandleTCP(address, uploadDir, DB)

	case "HTTP":
		HTTP.HandleHTTP(address, uploadDir, DB)

	case "GRPC":
		GRPC.HandleGRPC(address, uploadDir, DB)
	}

}
