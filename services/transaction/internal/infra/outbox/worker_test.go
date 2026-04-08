//go:build integration

package outbox

import (
	"context"
	"database/sql"
	"log"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/zap"
)

var testPool *pgxpool.Pool

type mockKafka struct{}

func (m *mockKafka) ProduceSync(ctx context.Context, rs ...*kgo.Record) kgo.ProduceResults {
	return kgo.ProduceResults{}
}

func TestProcessBatch_MarksSentAt(t *testing.T) {

	q := `
		INSERT INTO outbox (id, topic, key, payload)
		VALUES ($1, $2, $3, $4)
	`

	ctx := context.Background()

	id := uuid.New()
	_, err := testPool.Exec(ctx, q, id, "transaction.created", "key1", []byte(`{}`))
	require.NoError(t, err)

	m := &mockKafka{}
	w := NewWorker(testPool, m, time.Second, zap.NewNop())
	err = w.processBatch(ctx)
	require.NoError(t, err)

	var sentAt *time.Time
	q = `
		SELECT sent_at
		FROM outbox
		WHERE id = $1
	`

	err = w.db.QueryRow(ctx, q, id).Scan(&sentAt)
	require.NoError(t, err)
	require.NotNil(t, sentAt)
}

func TestMain(m *testing.M) {
	ctx := context.Background()

	container, err := tcpostgres.Run(ctx, "postgres:16",
		tcpostgres.WithDatabase("testdb"),
		tcpostgres.WithUsername("test"),
		tcpostgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").WithOccurrence(2),
		),
	)
	if err != nil {
		log.Fatal(err)
	}

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	db, err := sql.Open("pgx", connStr)
	if err != nil {
		log.Fatal(err)
	}

	err = goose.SetDialect("postgres")
	if err != nil {
		log.Fatal(err)
	}

	err = goose.Up(db, "../../../migrations")
	if err != nil {
		log.Fatal(err)
	}

	testPool, err = pgxpool.New(ctx, connStr)
	if err != nil {
		log.Fatal(err)
	}
	code := m.Run()
	container.Terminate(ctx)
	os.Exit(code)
}
