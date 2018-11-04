package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"log"
	"math/big"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/benguild/gke-grpc-example/protos"
	"github.com/golang/protobuf/ptypes/empty"
	"golang.org/x/net/http2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func main() {
	netInterfaces, err := net.Interfaces()

	if err != nil {
		log.Fatalf("failed to load net interfaces: %v", err)
		return
	}

	var ipAddress *net.IP

	for _, netInterface := range netInterfaces {
		ipAddresses, err := netInterface.Addrs()

		if err != nil {
			log.Fatalf("failed to load addresses for net interface: %v", err)
			return
		}

		for _, addr := range ipAddresses {
			switch v := addr.(type) {
			case *net.IPNet:
				ipAddress = &v.IP
			case *net.IPAddr:
				ipAddress = &v.IP
			}

			if ipAddress != nil {
				break
			}
		}

		if ipAddress != nil {
			break
		}
	}

	if ipAddress == nil {
		log.Fatalf("failed to discover IP address")
		return
	}

	issuer := pkix.Name{CommonName: ipAddress.String()}

	caCertificate := &x509.Certificate{
		SerialNumber:          big.NewInt(time.Now().UnixNano()),
		Subject:               issuer,
		Issuer:                issuer,
		SignatureAlgorithm:    x509.SHA512WithRSA,
		PublicKeyAlgorithm:    x509.ECDSA,
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(0, 0, 10),
		SubjectKeyId:          []byte{},
		BasicConstraintsValid: true,
		IsCA:        true,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
	}

	privateKey, _ := rsa.GenerateKey(rand.Reader, 4096)
	caCertificateBinary, err := x509.CreateCertificate(rand.Reader, caCertificate, caCertificate, &privateKey.PublicKey, privateKey)

	if err != nil {
		log.Fatalf("create cert failed: %v", err)
		return
	}

	caCertificateParsed, _ := x509.ParseCertificate(caCertificateBinary)

	certPool := x509.NewCertPool()
	certPool.AddCert(caCertificateParsed)

	tlsConfig := &tls.Config{
		ServerName: ipAddress.String(),
		Certificates: []tls.Certificate{{
			Certificate: [][]byte{caCertificateBinary},
			PrivateKey:  privateKey,
		}},
		RootCAs: certPool,
	}

	go func() {
		http.HandleFunc("/_ah/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, "Hello World.")
		})

		healthCheckPort, _ := strconv.Atoi(os.Getenv("PORT_HEALTHCHECK"))

		srv := &http.Server{
			Addr: fmt.Sprintf(":%d", healthCheckPort),
		}

		http2.ConfigureServer(srv, &http2.Server{})

		lis, err := tls.Listen("tcp", fmt.Sprintf(":%d", healthCheckPort), tlsConfig)

		if err != nil {
			log.Fatalf("failed to listen: %v", err)
			return
		}

		log.Fatal(srv.Serve(lis))
	}()

	grpcPort, _ := strconv.Atoi(os.Getenv("PORT_GRPC"))
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))

	if err != nil {
		log.Fatalf("failed to listen: %v", err)
		return
	}

	grpcServer := grpc.NewServer([]grpc.ServerOption{grpc.Creds(credentials.NewTLS(tlsConfig))}...)
	protos.RegisterExampleServiceServer(grpcServer, &service{})

	log.Fatal(grpcServer.Serve(lis))
}

type service struct{}

func (s *service) SayHello(ctx context.Context, req *empty.Empty) (*empty.Empty, error) {
	return &empty.Empty{}, nil
}
