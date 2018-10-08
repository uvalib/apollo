package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/uvalib/apollo/backend/internal/models"
)

func TestBadLookup(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Stub DB connection failed: %s", err)
	}
	defer mockDB.Close()
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	app := ApolloHandler{Version: "MOCK", DB: &models.DB{DB: sqlxDB}}

	mock.ExpectQuery("SELECT np.pid").WillReturnError(errors.New("Not Found"))

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.GET("/api/external/:pid", app.ExternalPIDLookup)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/external/uva-lib:bad", nil)
	router.ServeHTTP(w, req)

	// Check the status code is what we expect.
	if status := w.Code; status != http.StatusNotFound {
		t.Errorf("Wrong status code: got %v want %v", status, http.StatusOK)
	}
}

func TestLookup(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Stub DB connection failed: %s", err)
	}
	defer mockDB.Close()
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	app := ApolloHandler{Version: "MOCK", DB: &models.DB{DB: sqlxDB}}

	rows := sqlmock.NewRows([]string{"pid"}).AddRow("uva-an1")
	mock.ExpectQuery("SELECT np.pid").WillReturnRows(rows) // Query is CASE SENSITIVE

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.GET("/api/external/:pid", app.ExternalPIDLookup)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/external/uva-lib:100", nil)
	router.ServeHTTP(w, req)

	// Check the status code is what we expect.
	if status := w.Code; status != http.StatusOK {
		t.Errorf("Wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check the response body is what we expect.
	if strings.Contains(w.Body.String(), "uva-an1") == false {
		t.Errorf("Unexpected response: got [%s]. Does not include [%s]", w.Body.String(), "uva-an1")
	}
}
