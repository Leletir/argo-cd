package apiclient

import (
	"crypto/tls"
	"time"

	"github.com/argoproj/argo-cd/engine/util/misc"

	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	argogrpc "github.com/argoproj/argo-cd/util/grpc"
)

// Clientset represets repository server api clients
type Clientset interface {
	NewRepoServerClient() (misc.Closer, RepoServerServiceClient, error)
}

type clientSet struct {
	address        string
	timeoutSeconds int
}

func (c *clientSet) NewRepoServerClient() (misc.Closer, RepoServerServiceClient, error) {
	retryOpts := []grpc_retry.CallOption{
		grpc_retry.WithMax(3),
		grpc_retry.WithBackoff(grpc_retry.BackoffLinear(1000 * time.Millisecond)),
	}
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{InsecureSkipVerify: true})),
		grpc.WithStreamInterceptor(grpc_retry.StreamClientInterceptor(retryOpts...)),
		grpc.WithUnaryInterceptor(grpc_retry.UnaryClientInterceptor(retryOpts...))}
	if c.timeoutSeconds > 0 {
		opts = append(opts, grpc.WithUnaryInterceptor(argogrpc.WithTimeout(time.Duration(c.timeoutSeconds)*time.Second)))
	}
	conn, err := grpc.Dial(c.address, opts...)
	if err != nil {
		log.Errorf("Unable to connect to repository service with address %s", c.address)
		return nil, nil, err
	}
	return conn, NewRepoServerServiceClient(conn), nil
}

// NewRepoServerClientset creates new instance of repo server Clientset
func NewRepoServerClientset(address string, timeoutSeconds int) Clientset {
	return &clientSet{address: address, timeoutSeconds: timeoutSeconds}
}
