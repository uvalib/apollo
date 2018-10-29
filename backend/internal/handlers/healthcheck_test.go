package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/uvalib/apollo/backend/internal/models"
)

func TestHealthCheckFail(t *testing.T) {
	mockDB, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Stub DB connection failed: %s", err)
	}
	defer mockDB.Close()
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	app := Apollo{Version: "MOCK", DB: &models.DB{DB: sqlxDB}}

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.GET("/healthcheck", app.HealthCheck)

	req, _ := http.NewRequest("GET", "/healthcheck", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("Wrong status code: got %v want %v", status, http.StatusInternalServerError)
	}

	// Check the response body is what we expect.
	expected := `{"alive":"true","mysql":"false"}`
	if strings.TrimSpace(rr.Body.String()) != expected {
		t.Errorf("Unexpected response: got [%s] want [%s]", rr.Body.String(), expected)
	}
}

func TestHealthCheckPass(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Stub DB connection failed: %s", err)
	}
	defer mockDB.Close()
	mock.ExpectQuery("SELECT 1").WillReturnRows()
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	app := Apollo{Version: "MOCK", DB: &models.DB{DB: sqlxDB}}

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.GET("/healthcheck", app.HealthCheck)

	req, _ := http.NewRequest("GET", "/healthcheck", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check the response body is what we expect.
	expected := `{"alive":"true","mysql":"true"}`
	if strings.TrimSpace(rr.Body.String()) != expected {
		t.Errorf("Unexpected response: got [%s] want [%s]", rr.Body.String(), expected)
	}
}
