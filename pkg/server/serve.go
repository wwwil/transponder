package server

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var HTTPPort, HTTPSPort, GRPCPort uint16

func Serve(cmd *cobra.Command, args []string) {
	log.Println("Transponder is serving...")
	http.HandleFunc("/", handler)
	errs := make(chan error, 1)
	go serveHTTP(errs)
	go serveHTTPS(errs)
	go serveGRPC(errs)
	log.Fatal(<-errs)
}

func handler(writer http.ResponseWriter, request *http.Request) {
	hostname, err := os.Hostname()
	if err != nil {
		return
	}
	fmt.Fprintf(writer, "%s\n", hostname)
}

func serveHTTP(errs chan<- error) {
	errs <- http.ListenAndServe(fmt.Sprintf(":%d", HTTPPort), nil)
}

func serveHTTPS(errs chan<- error) {
	// Get the hostname.
	hostname, err := os.Hostname()
	if err != nil {
		return
	}

	// Generate a key and certificate.
	keyAndCert, err := GenerateKeyAndCert(KeyAndCertOpts{
		Host: hostname,
	})
	if err != nil {
		log.Fatalf("Failed to generate key and certificate to serve HTTPS: %v", err)
	}

	// Write the key to a temporary file.
	keyFile, err := ioutil.TempFile("", "key.*.pem")
	if err != nil {
		log.Fatalf("Failed to write temporary key file: %v", err)
	}
	keyFilePath := keyFile.Name()
	defer func() {
		err = os.Remove(keyFilePath)
		if err != nil {
			log.Printf("Unable to remove temporary key file: %q", err)
		}
	}()
	keyFile.Write(keyAndCert.Key)
	keyFile.Close()

	// Write the certificate to a temporary file.
	certFile, err := ioutil.TempFile("", "cert.*.pem")
	if err != nil {
		log.Fatalf("Failed to write temporary key file: %v", err)
	}
	certFilePath := certFile.Name()
	defer func() {
		err = os.Remove(certFilePath)
		if err != nil {
			log.Printf("Unable to remove temporary certificate file: %q", err)
		}
	}()
	certFile.Write(keyAndCert.Cert)
	certFile.Close()

	errs <- http.ListenAndServeTLS(fmt.Sprintf(":%d", HTTPSPort), certFilePath, keyFilePath, nil)
}

type greeterServer struct{}

// SayHello implements helloworld.GreeterServer.
func (s *greeterServer) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	log.Printf("Received: %v", in.Name)
	hostname, err := os.Hostname()
	if err != nil {
		log.Printf("Unable to get hostname %v", err)
	}
	if hostname != "" {
		grpc.SendHeader(ctx, metadata.Pairs("hostname", hostname))
	}
	return &pb.HelloReply{Message: "Hello " + in.Name}, nil
}

type healthServer struct{}

// Check is used for health checks.
func (s *healthServer) Check(ctx context.Context, in *healthpb.HealthCheckRequest) (*healthpb.HealthCheckResponse, error) {
	log.Printf("Handling Check request [%v]", in, ctx)
	return &healthpb.HealthCheckResponse{Status: healthpb.HealthCheckResponse_SERVING}, nil
}

// Watch is not implemented.
func (s *healthServer) Watch(in *healthpb.HealthCheckRequest, srv healthpb.Health_WatchServer) error {
	return status.Error(codes.Unimplemented, "Watch is not implemented")
}

func serveGRPC(errs chan<- error) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", GRPCPort))
	if err != nil {
		errs <- err
	}
	s := grpc.NewServer()
	pb.RegisterGreeterServer(s, &greeterServer{})
	healthpb.RegisterHealthServer(s, &healthServer{})
	if err := s.Serve(lis); err != nil {
		errs <- err
	}
}
