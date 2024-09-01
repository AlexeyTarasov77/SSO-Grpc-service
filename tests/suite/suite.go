package suite

import (
	"fmt"
	"net"
	"testing"

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
	Cfg *config.Config
	AuthClient ssov1.AuthClient
}

func New(t *testing.T) (context.Context, *Suite) {
	t.Helper()

	cfg := config.Load("../../config/local-tests.yaml")
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.Timeout)
	t.Cleanup(cancel)
	cc, err := grpc.NewClient(
		net.JoinHostPort(cfg.Server.Host, cfg.Server.Port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatal(err)
	}
	return ctx, &Suite{
		T: t,
		Cfg: cfg,
		AuthClient: ssov1.NewAuthClient(cc),
	}
}

func (st *Suite) NewTestStorage(t *testing.T) *postgres.Storage {
	t.Helper()
	storagePath := fmt.Sprintf(
		"postgresql://%s:%s@%s:%s/%s?sslmode=disable",
		st.Cfg.DB.User, st.Cfg.DB.Password, st.Cfg.DB.Host, st.Cfg.DB.Port, st.Cfg.DB.Name,
	)
	storage, err := postgres.New(storagePath)
	require.NoError(t, err)
	t.Cleanup(func() {
		storage.DB.Close()
	})
	return storage
}