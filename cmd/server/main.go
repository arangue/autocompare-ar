package main

import (
	"log"
	"net/http"
	"os"

	"github.com/arangue/autocompare-ar/internal/application"
	httpdelivery "github.com/arangue/autocompare-ar/internal/delivery/http"
	"github.com/arangue/autocompare-ar/internal/infrastructure/postgres"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Adapters (driven): infrastructure implements domain ports.
	vehicleRepo := postgres.NewVehicleRepository()
	_ = postgres.NewListingRepository()
	_ = postgres.NewPricingRepository()

	// Use cases (application).
	listBrands := application.NewListBrands(vehicleRepo)

	// Adapters (driving): HTTP handlers call use cases.
	handler := httpdelivery.NewHandler(listBrands)
	router := httpdelivery.NewRouter(handler)

	log.Printf("server listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
