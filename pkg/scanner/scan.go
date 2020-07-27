package scanner

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"
	"google.golang.org/grpc/metadata"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
	"crypto/tls"
	"gopkg.in/yaml.v2"
)

const (
	timeout = 5 * time.Second
)

var ConfigFilePath string

type Config struct {
	Servers []Server
}

type Server struct {
	Host  string
	Ports []Port
}

type Port struct {
	Number   uint16
	Protocol string
}

func Scan(cmd *cobra.Command, args []string) {
	configFile, err := os.Open(ConfigFilePath)
	if err != nil {
		log.Fatalf("Failed to load config file for scanner from: %s", ConfigFilePath)
	}
	defer configFile.Close()
	configFileData, err := ioutil.ReadAll(configFile)
	var config Config
	err = yaml.Unmarshal(configFileData, &config)
	if err != nil {
		log.Fatalf("Failed to unmarshal config file for scanner from: %s", ConfigFilePath)
	}
	log.Printf("Using config file from: %s\n", configFile.Name())
	for {
		for _, server := range config.Servers {
			for _, port := range server.Ports {
				switch strings.ToUpper(port.Protocol) {
				case "HTTP":
					scanHTTP(fmt.Sprintf("%s:%d", server.Host, port.Number))
				case "HTTPS":
					scanHTTPS(fmt.Sprintf("%s:%d", server.Host, port.Number))
				case "GRPC":
					scanGRPC(fmt.Sprintf("%s:%d", server.Host, port.Number))
				}
			}
			time.Sleep(timeout)
		}
	}
}

func scanHTTP(address string) {
	client := http.Client{
		Timeout: timeout,
	}
	resp, err := client.Get(fmt.Sprintf("http://%s", address))
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()
	body, err := httpReadBody(resp)
	if err != nil {
		log.Printf("HTTP: Error making request: %q", err)
	}
	log.Printf("HTTP: Successfully made request: %s", body)
}

func scanHTTPS(address string) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: tr,
		Timeout: timeout,
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("https://%s", address), nil)
	if err != nil {
		log.Println(err)
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()
	body, err := httpReadBody(resp)
	if err != nil {
		log.Printf("HTTPS: Error making request: %v", err)
	}
	log.Printf("HTTPS: Successfully made request: %s", body)
}

func httpReadBody(resp *http.Response) (string, error) {
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("got status code: %d", resp.StatusCode)
	}
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(bodyBytes), nil

}

func scanGRPC(address string) {
	//cert := flag.String("cert", "/data/cert.pem", "path to TLS certificate")
	//insecure := flag.Bool("insecure", false, "connect without TLS")

	// Set up a connection to the server.
	var conn *grpc.ClientConn
	var err error
	//if *insecure {
	conn, err = grpc.Dial(address, grpc.WithInsecure())
	//} else {
	//	tc, err := credentials.NewClientTLSFromFile(*cert, "")
	//	if err != nil {
	//		log.Fatalf("Failed to generate credentials %v", err)
	//	}
	//	conn, err = grpc.Dial(address, grpc.WithTransportCredentials(tc))
	//}
	if err != nil {
		log.Printf("did not connect: %v\n", err)
		return
	}
	defer conn.Close()
	c := pb.NewGreeterClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	var header metadata.MD
	r, err := c.SayHello(ctx, &pb.HelloRequest{}, grpc.Header(&header))
	if err != nil {
		log.Printf("gRPC: Error making request: %v", err)
		return
	}
	//hostname := "unknown"
	//if len(header["hostname"]) > 0 {
	//	hostname = header["hostname"][0]
	//}
	log.Printf("gRPC: Successfully made request: %s", r.Message)
}
