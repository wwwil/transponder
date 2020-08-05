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
				msg := fmt.Sprintf("%s:%d %s", server.Host, port.Number, port.Protocol)
				var resp string
				var err error
				switch strings.ToUpper(port.Protocol) {
				case "HTTP":
					resp, err = scanHTTP(fmt.Sprintf("%s:%d", server.Host, port.Number))
				case "HTTPS":
					resp, err = scanHTTPS(fmt.Sprintf("%s:%d", server.Host, port.Number))
				case "GRPC":
					resp, err = scanGRPC(fmt.Sprintf("%s:%d", server.Host, port.Number))
				}
				if err != nil {
					log.Printf("%s: Error making request: %q", msg, err)
					continue
				}
				log.Printf("%s: Successfully made request: %s", msg, resp)
			}
			time.Sleep(timeout)
		}
	}
}

func scanHTTP(address string) (string, error) {
	client := http.Client{
		Timeout: timeout,
	}
	resp, err := client.Get(fmt.Sprintf("http://%s", address))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := httpReadBody(resp)
	if err != nil {
		return "", err
	}
	return body, nil
}

func scanHTTPS(address string) (string, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: tr,
		Timeout: timeout,
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("https://%s", address), nil)
	if err != nil {
		return "", err
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := httpReadBody(resp)
	if err != nil {
		return "", err
	}
	return body, nil
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

func scanGRPC(address string) (string, error) {
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
		return "", err
	}
	defer conn.Close()
	c := pb.NewGreeterClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	var header metadata.MD
	r, err := c.SayHello(ctx, &pb.HelloRequest{}, grpc.Header(&header))
	if err != nil {
		return "", err
	}
	//hostname := "unknown"
	//if len(header["hostname"]) > 0 {
	//	hostname = header["hostname"][0]
	//}
	return r.Message, nil
}
