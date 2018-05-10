package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/julienschmidt/httprouter"
	"github.com/uvalib/apollo/backend/internal/models"
)

func TestCollectionsIndex(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Stub DB connection failed: %s", err)
	}
	defer mockDB.Close()
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	app := ApolloHandler{Version: "MOCK", DB: &models.DB{sqlxDB}}

	req, _ := http.NewRequest("GET", "/api/collections", nil)
	rr := httptest.NewRecorder()

	rows := sqlmock.NewRows([]string{"id", "pid"}).AddRow(1, "an666")
	mock.ExpectQuery("select id,pid from nodes").WillReturnRows(rows)
	titleRows := sqlmock.NewRows([]string{"value"}).AddRow("test")
	mock.ExpectQuery("select value from nodes").WithArgs(1, 2).WillReturnRows(titleRows)

	app.CollectionsIndex(rr, req, httprouter.Params{})

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check the response body is what we expect.
	expected := `[{"pid":"an666","title":"test"}]`
	if strings.TrimSpace(rr.Body.String()) != expected {
		t.Errorf("Unexpected response: got [%s] want [%s]", rr.Body.String(), expected)
	}
}

func TestBadCollectionShow(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Stub DB connection failed: %s", err)
	}
	defer mockDB.Close()
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	app := ApolloHandler{Version: "MOCK", DB: &models.DB{sqlxDB}}
	mock.ExpectQuery("SELECT").WillReturnError(errors.New("Collection not found"))

	req, _ := http.NewRequest("GET", "/api/collection", nil)
	rr := httptest.NewRecorder()

	params := httprouter.Params{httprouter.Param{"pid", "bad123"}}
	app.CollectionsShow(rr, req, params)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("Wrong status code: got %v want %v", status, http.StatusNotFound)
	}
}

func TestCollectionShow(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Stub DB connection failed: %s", err)
	}
	defer mockDB.Close()
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	app := ApolloHandler{Version: "MOCK", DB: &models.DB{sqlxDB}}

	tgt := "uva-an1"
	rowsPid := sqlmock.NewRows([]string{"id"}).AddRow(1)
	rows := sqlmock.NewRows([]string{"n.id", "n.parent_id", "n.pid", "n.value", "n.created_at", "n.updated_at", "nn.pid", "nn.value", "nn.controlled_vocab"}).
		AddRow(1, nil, tgt, "woof", nil, nil, "uva-ann1", "collection", 0)
	mock.ExpectQuery("select id from nodes").WillReturnRows(rowsPid)
	mock.ExpectQuery("SELECT n.id").WithArgs(1, 1).WillReturnRows(rows)

	req, _ := http.NewRequest("GET", "/api/collection", nil)
	rr := httptest.NewRecorder()

	params := httprouter.Params{httprouter.Param{"pid", tgt}}
	app.CollectionsShow(rr, req, params)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Wrong status code: got %v want %v", status, http.StatusOK)
	}

	if strings.Contains(strings.TrimSpace(rr.Body.String()), "uva-an1") == false {
		t.Errorf("Response %s does not contain searched PID: %s", rr.Body.String(), tgt)
	}
}
