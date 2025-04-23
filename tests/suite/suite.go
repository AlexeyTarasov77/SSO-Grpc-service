package suite

import (
	"net"
	"testing"
	"time"

	ssov1 "github.com/AlexeySHA256/protos/gen/go/sso"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"sso.service/internal/config"
	"sso.service/internal/storage/postgres"
)

type Suite struct {
	*testing.T
	Cfg               *config.Config
	AuthClient        ssov1.AuthClient
	PermissionsClient ssov1.PermissionsClient
}

func New(t *testing.T) *Suite {
	t.Helper()

	cfg := config.MustLoad("../../../config/local-tests.yaml")
	conn, err := grpc.NewClient(
		net.JoinHostPort(cfg.Server.Host, cfg.Server.Port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatal(err)
	}
	return &Suite{
		T:                 t,
		Cfg:               cfg,
		AuthClient:        ssov1.NewAuthClient(conn),
		PermissionsClient: ssov1.NewPermissionsClient(conn),
	}
}

func (self *Suite) NewTestStorage() *postgres.Storage {
	self.T.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	storage, err := postgres.New(ctx, self.Cfg.DB.Dsn)
	require.NoError(self.T, err)
	self.T.Cleanup(func() {
		storage.DB.Close()
	})
	return storage
}
