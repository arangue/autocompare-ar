package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/arangue/autocompare-ar/internal/application"
	httpdelivery "github.com/arangue/autocompare-ar/internal/delivery/http"
	"github.com/arangue/autocompare-ar/internal/infrastructure/postgres"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		slog.Error("DATABASE_URL is not set")
		os.Exit(1)
	}

	poolCfg, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		slog.Error("failed to parse database URL", "error", err)
		os.Exit(1)
	}
	poolCfg.MaxConns = 25
	poolCfg.MinConns = 2
	poolCfg.MaxConnLifetime = 30 * time.Minute
	poolCfg.MaxConnIdleTime = 5 * time.Minute

	pool, err := pgxpool.NewWithConfig(context.Background(), poolCfg)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := pool.Ping(context.Background()); err != nil {
		slog.Error("database is unreachable", "error", err)
		os.Exit(1)
	}

	if err := runMigrations(dbURL); err != nil {
		slog.Error("migrations failed", "error", err)
		os.Exit(1)
	}

	vehicleRepo := postgres.NewVehicleRepository(pool)
	listingRepo := postgres.NewListingRepository(pool)
	pricingRepo := postgres.NewPricingRepository(pool)
	listBrands := application.NewListBrands(vehicleRepo)
	listModelsByBrand := application.NewListModelsByBrand(vehicleRepo)
	searchTrims := application.NewSearchTrims(vehicleRepo)
	getTrim := application.NewGetTrimService(vehicleRepo)
	listListings := application.NewListListings(listingRepo)
	getMarketSummary := application.NewGetMarketSummary(listingRepo)
	listReferences := application.NewListReferences(pricingRepo)
	assessDeal := application.NewAssessDeal(listingRepo)
	compareTrims := application.NewCompareTrims(getTrim, getMarketSummary)
	handler := httpdelivery.NewHandler(listBrands, listModelsByBrand, searchTrims, getTrim, listListings, getMarketSummary, listReferences, assessDeal, compareTrims, pool)
	router := httpdelivery.NewRouter(handler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("server started", "port", port)
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down server")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("graceful shutdown failed", "error", err)
		os.Exit(1)
	}

	slog.Info("server stopped")
}
