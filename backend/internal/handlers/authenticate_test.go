package handlers

import (
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/uvalib/apollo/backend/internal/models"
)

func TestMissingAuthenticate(t *testing.T) {
	log.Printf("Testing no credentials of Authenticate API...")
	mockDB, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Stub DB connection failed: %s", err)
	}
	defer mockDB.Close()
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	app := ApolloHandler{Version: "MOCK", DB: &models.DB{sqlxDB}}

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.GET("/api/authenticate", app.Authenticate)

	req, _ := http.NewRequest("GET", "/api/authenticate", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusForbidden {
		t.Errorf("Wrong status code: got %v want %v", status, http.StatusForbidden)
	}
}

func TestGoodAuthenticate(t *testing.T) {
	log.Printf("Testing Good use of Authenticate API....")
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Stub DB connection failed: %s", err)
	}
	defer mockDB.Close()
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	app := ApolloHandler{Version: "MOCK", DB: &models.DB{sqlxDB}, DevAuthUser: "lf6f"}

	rows := sqlmock.NewRows([]string{"id", "computing_id", "first_name", "last_name", "email"}).
		AddRow(1, "lf6f", "Lou", "Foster", "lf6f")
	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.GET("/api/authenticate", app.Authenticate)

	req, _ := http.NewRequest("GET", "/api/authenticate", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Wrong status code: got %v want %v", status, http.StatusOK)
	}
}

func TestBadAuthenticate(t *testing.T) {
	log.Printf("Testing Good use of Authenticate API....")
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Stub DB connection failed: %s", err)
	}
	defer mockDB.Close()
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	app := ApolloHandler{Version: "MOCK", DB: &models.DB{sqlxDB}, DevAuthUser: "wrong"}

	mock.ExpectQuery("SELECT").WillReturnError(errors.New("You are not authorized to access this site"))

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.GET("/api/authenticate", app.Authenticate)

	req, _ := http.NewRequest("GET", "/api/authenticate", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusForbidden {
		t.Errorf("Wrong status code: got %v want %v", status, http.StatusOK)
	}
}
