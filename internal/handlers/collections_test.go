package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/julienschmidt/httprouter"
	"github.com/uvalib/apollo/internal/models"
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

	rows := sqlmock.NewRows([]string{"pid"}).AddRow("an666")
	mock.ExpectQuery("select pid from nodes").WillReturnRows(rows)

	app.CollectionsIndex(rr, req, httprouter.Params{})

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check the response body is what we expect.
	expected := `["an666"]`
	if strings.TrimSpace(rr.Body.String()) != expected {
		t.Errorf("Unexpected response: got [%s] want [%s]", rr.Body.String(), expected)
	}
}

func TestBadCollectionShow(t *testing.T) {
	// mockDB, mock, err := sqlmock.New()
	// if err != nil {
	// 	t.Fatalf("Stub DB connection failed: %s", err)
	// }
	// defer mockDB.Close()
	// sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	// app := ApolloHandler{Version: "MOCK", DB: &models.DB{sqlxDB}}
	// mock.ExpectQuery("SELECT").WillReturnError(errors.New("Collection not found"))
	//
	// req, _ := http.NewRequest("GET", "/api/collection", nil)
	// rr := httptest.NewRecorder()
	//
	// params := httprouter.Params{httprouter.Param{"pid", "bad123"}}
	// app.CollectionsShow(rr, req, params)
	//
	// // Check the status code is what we expect.
	// if status := rr.Code; status != http.StatusNotFound {
	// 	t.Errorf("Wrong status code: got %v want %v", status, http.StatusOK)
	// }
}
