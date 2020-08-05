package server

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/wwwil/transponder/pkg/version"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
)

var (
	HTTPPort, HTTPSPort, GRPCPort uint16
	Hostname                      string
)

func Serve(cmd *cobra.Command, args []string) {
	log.Println(version.ToString(false))
	log.Println("Transponder server is starting.")
	var err error
	Hostname, err = os.Hostname()
	if err != nil {
		log.Fatalf("Unable to determine hostname: %v", err)
	}
	log.Printf("Hostname is %s.", Hostname)
	http.HandleFunc("/", handler)
	errs := make(chan error, 1)
	go serveHTTP(errs)
	go serveHTTPS(errs)
	go serveGRPC(errs)
	for {
		log.Println(<-errs)
	}
}

func handler(writer http.ResponseWriter, request *http.Request) {
	_, err := fmt.Fprintf(writer, "Hello from %s\n", Hostname)
	if err != nil {
		log.Printf("Unable to handle request: %v", err)
	}
}

func serveHTTP(errs chan<- error) {
	log.Printf("Serving HTTP on port %d.", HTTPPort)
	errs <- http.ListenAndServe(fmt.Sprintf(":%d", HTTPPort), nil)
}

func serveHTTPS(errs chan<- error) {
	log.Println("Generating certificate to serve HTTPS.")
	// Generate a key and certificate.
	keyAndCert, err := GenerateKeyAndCert(KeyAndCertOpts{
		Host: Hostname,
	})
	if err != nil {
		log.Printf("Failed to generate key and certificate to serve HTTPS: %v", err)
		return
	}

	// Write the key to a temporary file.
	keyFile, err := ioutil.TempFile("", "key.*.pem")
	if err != nil {
		log.Printf("Failed to write temporary key file to serve HTTPS: %v", err)
		return
	}
	keyFilePath := keyFile.Name()
	defer func() {
		err = os.Remove(keyFilePath)
		if err != nil {
			log.Printf("Unable to remove temporary key file used to serve HTTPS: %q", err)
		}
	}()
	keyFile.Write(keyAndCert.Key)
	keyFile.Close()

	// Write the certificate to a temporary file.
	certFile, err := ioutil.TempFile("", "cert.*.pem")
	if err != nil {
		log.Printf("Failed to write temporary key file to serve HTTPS: %v", err)
		return
	}
	certFilePath := certFile.Name()
	defer func() {
		err = os.Remove(certFilePath)
		if err != nil {
			log.Printf("Unable to remove temporary certificate file used to serve HTTPS: %q", err)
		}
	}()
	certFile.Write(keyAndCert.Cert)
	certFile.Close()

	log.Printf("Serving HTTPS on port %d.", HTTPSPort)
	errs <- http.ListenAndServeTLS(fmt.Sprintf(":%d", HTTPSPort), certFilePath, keyFilePath, nil)
}

type greeterServer struct{}

// SayHello implements helloworld.GreeterServer.
func (s *greeterServer) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	grpc.SendHeader(ctx, metadata.Pairs("hostname", Hostname))
	return &pb.HelloReply{Message: fmt.Sprintf("Hello from %s", Hostname)}, nil
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
	log.Printf("Serving gRPC on port %d.", GRPCPort)
	//errs <- s.Serve(lis)
	if err := s.Serve(lis); err != nil {
		errs <- err
	}
}
