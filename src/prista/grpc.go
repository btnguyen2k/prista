package prista

import (
	"context"
	"fmt"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"
	"io"
	"log"
	pb "main/src/grpc"
	"net"
	"strings"
	"sync"
)

// initialize and start gRPC server
func initGrpcServer(wg *sync.WaitGroup) bool {
	listenPort := AppConfig.GetInt32("server.grpc.listen_port", 0)
	if listenPort <= 0 {
		log.Println("No valid [server.grpc.listen_port] configured, gRPC server is disabled.")
		return false
	}
	listenAddr := AppConfig.GetString("server.grpc.listen_addr", "127.0.0.1")

	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", listenAddr, listenPort))
	if err != nil {
		log.Printf("Failed to listen gRPC: %v", err)
		return false
	}
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterPLogCollectorServiceServer(grpcServer, &PLogCollectorServiceServer{})
	log.Printf("Starting [%s] gRPC server on [%s:%d]...\n", AppConfig.GetString("app.name")+" v"+AppConfig.GetString("app.version"), listenAddr, listenPort)
	go func() {
		err := grpcServer.Serve(lis)
		if err != nil {
			log.Println(err)
		}
		wg.Done()
	}()
	return true
}

// PLogCollectorServiceServer is gRPC server to handle log request
type PLogCollectorServiceServer struct {
}

// Ping implements PLogCollectorServiceServer.Ping
func (server *PLogCollectorServiceServer) Ping(context.Context, *empty.Empty) (*empty.Empty, error) {
	return &empty.Empty{}, nil
}

// Ping implements PLogCollectorServiceServer.Log
func (server *PLogCollectorServiceServer) Log(_ context.Context, msg *pb.PLogMessage) (*pb.PLogResult, error) {
	category := strings.TrimSpace(msg.Category)
	message := strings.TrimSpace(msg.Message)
	if category == "" || message == "" {
		return &pb.PLogResult{
			Status:     400,
			NumSuccess: 0,
			Message:    "Missing parameter [category] and/or [message]",
		}, nil
	}
	payload := strings.ToLower(category) + "\t" + message
	if err := handleIncomingMessage([]byte(payload)); err != nil {
		return &pb.PLogResult{
			Status:     500,
			NumSuccess: 0,
			Message:    err.Error(),
		}, nil
	}
	return &pb.PLogResult{
		Status:     200,
		NumSuccess: 1,
		Message:    "Ok",
	}, nil
}

// Ping implements PLogCollectorServiceServer.LogStream
func (server *PLogCollectorServiceServer) LogStream(msgs pb.PLogCollectorService_LogStreamServer) error {
	result := &pb.PLogResult{
		NumSuccess: 0,
	}
	for {
		msg, err := msgs.Recv()
		if err == io.EOF {
			result.Status = 200
			result.Message = "Ok"
			return msgs.SendAndClose(result)
		}
		if err != nil {
			result.Status = 500
			result.Message = err.Error()
			return msgs.SendAndClose(result)
		}
		category := strings.TrimSpace(msg.Category)
		message := strings.TrimSpace(msg.Message)
		if category == "" || message == "" {
			result.Status = 400
			result.Message = "Missing parameter [category] and/or [message]"
			return msgs.SendAndClose(result)
		}
		payload := strings.ToLower(category) + "\t" + message
		if err := handleIncomingMessage([]byte(payload)); err != nil {
			result.Status = 500
			result.Message = err.Error()
			return msgs.SendAndClose(result)
		}
		result.NumSuccess++
	}
}
