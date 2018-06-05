package handlers

import (
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/julienschmidt/httprouter"
	"github.com/uvalib/apollo/backend/internal/models"
)

func TestBadItemShow(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Stub DB connection failed: %s", err)
	}
	defer mockDB.Close()
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	app := ApolloHandler{Version: "MOCK", DB: &models.DB{sqlxDB}}
	mock.ExpectQuery("SELECT").WillReturnError(errors.New("Item not found"))

	req, _ := http.NewRequest("GET", "/api/items", nil)
	rr := httptest.NewRecorder()

	params := httprouter.Params{httprouter.Param{"pid", "bad123"}}
	app.ItemShow(rr, req, params)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusNotFound {
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
	app := ApolloHandler{Version: "MOCK", DB: &models.DB{sqlxDB}}

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

	req, _ := http.NewRequest("GET", "/api/items", nil)
	rr := httptest.NewRecorder()

	params := httprouter.Params{httprouter.Param{"pid", tgt}}
	app.ItemShow(rr, req, params)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Wrong status code: got %v want %v", status, http.StatusOK)
	}

	if strings.Contains(strings.TrimSpace(rr.Body.String()), tgt) == false {
		t.Errorf("Response %s does not contain searched PID: %s", rr.Body.String(), tgt)
	}
}
