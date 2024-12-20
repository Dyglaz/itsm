// main.go
package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"itsm/api/auth"
	"itsm/api/dashboard"
	"itsm/api/incidents"
	"itsm/api/messenger"
	"itsm/api/services"
	"itsm/models"
	_ "itsm/session"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

func startServer(port string, db *gorm.DB) {
	r := mux.NewRouter()
	auth.SetupRoutes(r, db)
	dashboard.SetupRoutes(r, db)
	services.SetupRoutes(r, db)
	incidents.SetupRoutes(r, db)
	messenger.SetupRoutes(r, db)

	fs := http.FileServer(http.Dir("./templates"))
	r.PathPrefix("/templates/").Handler(http.StripPrefix("/templates/", fs))

	log.Printf("Сервер запущен на :%s\n", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatal("Ошибка при запуске сервера:", err)
	}
}

func checkEnvVariables() {
	requiredEnvVars := []string{"DB_USER",
		"DB_HOST",
		"DB_NAME",
		"PORT1",
		"PORT2",
		"USE_2_SERVERS"}

	for _, envVar := range requiredEnvVars {
		value, exists := os.LookupEnv(envVar)
		if !exists || value == "" {
			log.Fatalf("Error: Environment variable %s is required but not set or empty.", envVar)
		}
	}
}

func getEnv(name string) string {
	return os.Getenv(name)
}

func startGoroutines(db *gorm.DB) {
	go startServer(getEnv("PORT1"), db)

	use2servers, err := strconv.ParseBool(getEnv("USE_2_SERVERS"))
	if err != nil {
		log.Fatalf("Error converting .env var USE_2_SERVERS to boolean: %v", err)
	}

	if use2servers {
		go startServer(getEnv("PORT2"), db)
	}

	select {}
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	checkEnvVariables()

	dbUser := getEnv("DB_USER")
	dbPass := getEnv("DB_PASS")
	dbHost := getEnv("DB_HOST")
	dbName := getEnv("DB_NAME")
	loc := url.QueryEscape("Europe/Moscow")

	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true&loc=%s", dbUser, dbPass, dbHost, dbName, loc)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	err = db.AutoMigrate(&models.User{}, &models.Service{}, &models.Message{},
		&models.Dialog{}, &models.Incident{})
	if err != nil {
		log.Fatal(err)
	}

	startGoroutines(db)
}
