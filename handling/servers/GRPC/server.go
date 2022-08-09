package GRPC

import (
	"encoding/json"
	"fmt"
	"log"
	"net"

	"io"

	"os"
	process "pcap_receiver/handling/process"
	"pcap_receiver/proto/storage"
	"time"

	"google.golang.org/grpc"

	uploadpb "pcap_receiver/proto/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var DB process.Saver

func HandleGRPC(addr, dirName string, DBase process.Saver) {
	DB = DBase

	lis, err := net.Listen("tcp", addr)
	fmt.Println("GRPC server has started")
	if err != nil {
		log.Fatal(err)
	}
	defer lis.Close()

	uplSrv := NewServer(storage.New(dirName), dirName)

	rpcSrv := grpc.NewServer()

	uploadpb.RegisterUploadServiceServer(rpcSrv, uplSrv)
	log.Fatal(rpcSrv.Serve(lis))
}

var dirName string

type Server struct {
	uploadpb.UnimplementedUploadServiceServer
	storage storage.Manager
	dir     string
}

func NewServer(storage storage.Manager, dir string) Server {
	return Server{
		storage: storage,
		dir:     dir,
	}
}

func (s Server) Upload(stream uploadpb.UploadService_UploadServer) error {

	name := s.dir + "/" + time.Now().Format("01-02-2002-123456465") + ".pcapng"
	_, err := os.OpenFile(name, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0600)
	f := storage.NewFile(name)
	if err != nil {
		return err
	}

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			if err := s.storage.Store(f); err != nil {
				return status.Error(codes.Internal, err.Error())
			}

			counter, err := process.Handle{}.Process(name)
			if err != nil {
				return err
			}

			b, err := json.Marshal(&counter)

			if err != nil {
				return err
			}
			DB.SaveToDB(counter, name)
			return stream.SendAndClose(&uploadpb.UploadResponse{Name: string(b)})
		}
		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}

		if err = f.Write(req.GetChunk()); err != nil {
			return status.Error(codes.Internal, err.Error())
		}

	}

}
