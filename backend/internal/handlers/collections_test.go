package handlers

import (
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/uvalib/apollo/backend/internal/models"
)

func TestCollectionsIndex(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Stub DB connection failed: %s", err)
	}
	defer mockDB.Close()
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	app := ApolloHandler{Version: "MOCK", DB: &models.DB{DB: sqlxDB}}

	rows := sqlmock.NewRows([]string{"id", "pid"}).AddRow(1, "an666")
	mock.ExpectQuery("select id,pid from nodes").WillReturnRows(rows)
	titleRows := sqlmock.NewRows([]string{"value"}).AddRow("test")
	mock.ExpectQuery("select value from nodes").WithArgs(1, 2).WillReturnRows(titleRows)

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.GET("/api/collections", app.CollectionsIndex)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/collections", nil)
	router.ServeHTTP(w, req)

	// Check the status code is what we expect.
	if status := w.Code; status != http.StatusOK {
		t.Errorf("Wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check the response body is what we expect.
	expected := `[{"pid":"an666","title":"test"}]`
	if strings.TrimSpace(w.Body.String()) != expected {
		t.Errorf("Unexpected response: got [%s] want [%s]", w.Body.String(), expected)
	}
}

func TestBadCollectionShow(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Stub DB connection failed: %s", err)
	}
	defer mockDB.Close()
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	app := ApolloHandler{Version: "MOCK", DB: &models.DB{DB: sqlxDB}}
	mock.ExpectQuery("SELECT").WillReturnError(errors.New("Collection not found"))

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.GET("/api/collections/:pid", app.CollectionsShow)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/collections/bad123", nil)
	router.ServeHTTP(w, req)

	// Check the status code is what we expect.
	if status := w.Code; status != http.StatusNotFound {
		t.Errorf("Wrong status code: got %v want %v", status, http.StatusNotFound)
	}
}

func TestCollectionShow(t *testing.T) {
	log.Printf("=== TestCollectionsShow")
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Stub DB connection failed: %s", err)
	}
	defer mockDB.Close()
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	app := ApolloHandler{Version: "MOCK", DB: &models.DB{DB: sqlxDB}}

	tgt := "uva-an1"
	rowsPid := sqlmock.NewRows([]string{"id"}).AddRow(1)
	mock.ExpectQuery("select id from nodes").WillReturnRows(rowsPid)

	rows := sqlmock.NewRows([]string{
		"n.id", "n.parent_id", "n.ancestry", "n.sequence", "n.pid", "n.value", "n.created_at", "n.updated_at",
		"nt.pid", "nt.value", "nt.controlled_vocab", "nt.container"}).
		AddRow(1, nil, nil, 0, tgt, "woof", nil, nil, "uva-ann1", "collection", 0, 1)
	mock.ExpectQuery("SELECT n.id").WithArgs(1).WillReturnRows(rows)

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.GET("/api/collections/:pid", app.CollectionsShow)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/collections/uva-an1", nil)
	router.ServeHTTP(w, req)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}

	// Check the status code is what we expect.
	if status := w.Code; status != http.StatusOK {
		t.Errorf("Wrong status code: got %v want %v", status, http.StatusOK)
	}

	if strings.Contains(strings.TrimSpace(w.Body.String()), "uva-an1") == false {
		t.Errorf("Response %s does not contain searched PID: %s", w.Body.String(), tgt)
	}
}
