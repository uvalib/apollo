package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/julienschmidt/httprouter"
	"github.com/uvalib/apollo/backend/internal/models"
)

func TestNamesIndex(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Stub DB connection failed: %s", err)
	}
	defer mockDB.Close()
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	app := ApolloHandler{Version: "MOCK", DB: &models.DB{sqlxDB}}

	rows := sqlmock.NewRows([]string{"pid", "value"}).
		AddRow("uva-ann1", "collection").AddRow("uva-ann2", "title")
	mock.ExpectQuery("select").WillReturnRows(rows) // Query is CASE SENSITIVE

	req, _ := http.NewRequest("GET", "/api/names", nil)
	rr := httptest.NewRecorder()
	app.NamesIndex(rr, req, httprouter.Params{})

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check the response body is what we expect.
	expected := `[{"pid": uva-ann1, "name": "collection"},{"pid": uva-ann2, "name": "title"}]`
	if strings.TrimSpace(rr.Body.String()) != expected {
		t.Errorf("Unexpected response: got [%s] want [%s]", rr.Body.String(), expected)
	}
}
