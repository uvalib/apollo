package handlers

import (
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/uvalib/apollo/backend/internal/models"
)

func TestBadItemShow(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Stub DB connection failed: %s", err)
	}
	defer mockDB.Close()
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	app := ApolloHandler{Version: "MOCK", DB: &models.DB{DB: sqlxDB}}
	mock.ExpectQuery("SELECT").WillReturnError(errors.New("Item not found"))

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.GET("/api/items/:pid", app.ItemShow)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/items/bad123", nil)
	router.ServeHTTP(w, req)

	// Check the status code is what we expect.
	if status := w.Code; status != http.StatusNotFound {
		t.Errorf("Wrong status code: got %v want %v", status, http.StatusNotFound)
	}
}

func TestItemShow(t *testing.T) {
	log.Printf("=== TestItemShow")
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Stub DB connection failed: %s", err)
	}
	defer mockDB.Close()
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	app := ApolloHandler{Version: "MOCK", DB: &models.DB{DB: sqlxDB}}

	tgt := "uva-an10"

	rowsPid := sqlmock.NewRows([]string{"id"}).AddRow(10)
	mock.ExpectQuery("select id from nodes").WithArgs(tgt).WillReturnRows(rowsPid)

	rows := sqlmock.NewRows([]string{
		"n.id", "n.parent_id", "n.ancestry", "n.sequence", "n.pid", "n.value", "n.created_at", "n.updated_at",
		"nt.pid", "nt.name", "nt.controlled_vocab", "nt.container"}).
		AddRow(1, nil, nil, 0, "uva-an1", "PARENT", nil, nil, "uva-ant1", "collection", 0, 1).
		AddRow(10, 1, "1", 0, tgt, "bark", nil, nil, "uva-ant7", "item", 0, 0)
	mock.ExpectQuery("SELECT n.id").WithArgs(10).WillReturnRows(rows)

	collectionRows := sqlmock.NewRows([]string{
		"n.id", "n.parent_id", "n.ancestry", "n.sequence", "n.pid", "n.value", "n.created_at", "n.updated_at",
		"nt.pid", "nt.name", "nt.controlled_vocab", "nt.container"}).
		AddRow(1, nil, nil, 0, "uva-an1", "PARENT", nil, nil, "uva-ant1", "collection", 0, 1)
	mock.ExpectQuery("SELECT n.id").WithArgs(1).WillReturnRows(collectionRows)

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.GET("/api/items/:pid", app.ItemShow)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/items/uva-an10", nil)
	router.ServeHTTP(w, req)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}

	// Check the status code is what we expect.
	if status := w.Code; status != http.StatusOK {
		t.Errorf("Wrong status code: got %v want %v", status, http.StatusOK)
	}

	if strings.Contains(strings.TrimSpace(w.Body.String()), tgt) == false {
		t.Errorf("Response %s does not contain searched PID: %s", w.Body.String(), tgt)
	}
}
