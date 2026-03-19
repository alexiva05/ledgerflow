//go:build integration

package postgres

import (
	"context"
	"ledgerflow/services/account/internal/app"
	"ledgerflow/services/account/internal/domain"
	"log"
	"os"
	"testing"

	"database/sql"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib" // регистрирует драйвер "pgx"
	"github.com/pressly/goose/v3"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var testPool *pgxpool.Pool

// func setupTestDB(t *testing.T) *pgxpool.Pool {

// 	t.Helper()
// 	ctx := context.Background()

// 	container, err := tcpostgres.Run(ctx, "postgres:16",
// 		tcpostgres.WithDatabase("testdb"),
// 		tcpostgres.WithUsername("test"),
// 		tcpostgres.WithPassword("test"),
// 		testcontainers.WithWaitStrategy(
// 			wait.ForLog("database system is ready to accept connections").WithOccurrence(2),
// 		),
// 	)
// 	require.NoError(t, err)

// 	t.Cleanup(func() {
// 		container.Terminate(ctx)
// 	})

// 	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
// 	require.NoError(t, err)

// 	db, err := sql.Open("pgx", connStr)
// 	require.NoError(t, err)
// 	t.Cleanup(func() {
// 		db.Close()
// 	})

// 	err = goose.SetDialect("postgres")
// 	require.NoError(t, err)

// 	err = goose.Up(db, "../../../migrations")
// 	require.NoError(t, err)

// 	pool, err := pgxpool.New(ctx, connStr)
// 	require.NoError(t, err)

// 	return pool
// }

func TestAccountRepo_Create(t *testing.T) {
	// pool := setupTestDB(t)
	// repo := NewAccountRepo(pool)
	repo := NewAccountRepo(testPool)
	ctx := context.Background()

	account := domain.Account{
		ID: uuid.New(), 
		Owner: uuid.New(),
		Currency: "USD", 
		Status: domain.StatusActive,
	}

	err := repo.Create(ctx, account)
	require.NoError(t, err)

	got, err := repo.GetByID(ctx, account.ID)
	require.NoError(t, err)

	require.Equal(t, account.ID, got.ID)
	require.Equal(t, account.Currency, got.Currency)
	require.Equal(t, account.Status, got.Status)
	require.Equal(t, account.Owner, got.Owner)
}

func TestAccountRepo_GetByID_NotFound(t *testing.T) {
	// pool := setupTestDB(t)
	// repo := NewAccountRepo(pool)
	repo := NewAccountRepo(testPool)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, uuid.New())
	require.ErrorIs(t, err, domain.ErrNotFound)
}

func TestAccountService_GetBalance(t *testing.T) {
	// pool := setupTestDB(t)
	// repo := NewAccountRepo(pool)
	repo := NewAccountRepo(testPool)
	ctx := context.Background()

	account := domain.Account{
		ID: uuid.New(),
		Owner: uuid.New(),
		Currency: "USD",
		Status: domain.StatusActive,
	}

	err := repo.Create(ctx, account)
	require.NoError(t, err)

	service := app.NewAccountService(repo)
	balance, err := service.GetBalance(ctx, account.ID)
	require.NoError(t, err)

	require.True(t, decimal.Zero.Equal(balance))
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