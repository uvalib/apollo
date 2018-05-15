package handlers

import (
	"errors"
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
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Stub DB connection failed: %s", err)
	}
	defer mockDB.Close()
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	app := ApolloHandler{Version: "MOCK", DB: &models.DB{sqlxDB}}

	tgt := "uva-an10"
	rowsPid := sqlmock.NewRows([]string{"id"}).AddRow(10)
	rows := sqlmock.NewRows([]string{
		"n.id", "n.parent_id", "n.pid", "n.value", "n.created_at", "n.updated_at",
		"nn.pid", "nn.value", "nn.controlled_vocab"}).
		AddRow(10, 1, tgt, "bark", nil, nil, "uva-ann7", "item", 0)
	mock.ExpectQuery("select id from nodes").WillReturnRows(rowsPid)
	mock.ExpectQuery("SELECT n.id").WithArgs(10).WillReturnRows(rows)

	ancestry := sqlmock.NewRows([]string{"ancestry"}).AddRow("1/10")
	mock.ExpectQuery("select ancestry").WillReturnRows(ancestry)

	collectionRows := sqlmock.NewRows([]string{
		"n.id", "n.parent_id", "n.pid", "n.value", "n.created_at", "n.updated_at",
		"nn.pid", "nn.value", "nn.controlled_vocab"}).
		AddRow(1, nil, tgt, "woof", nil, nil, "uva-ann1", "collection", 0)
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
