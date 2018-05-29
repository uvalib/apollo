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

func TestTypesIndex(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Stub DB connection failed: %s", err)
	}
	defer mockDB.Close()
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	app := ApolloHandler{Version: "MOCK", DB: &models.DB{sqlxDB}}

	rows := sqlmock.NewRows([]string{"pid", "name"}).
		AddRow("uva-ann1", "collection").AddRow("uva-ann2", "title")
	mock.ExpectQuery("select").WillReturnRows(rows) // Query is CASE SENSITIVE

	req, _ := http.NewRequest("GET", "/api/types", nil)
	rr := httptest.NewRecorder()
	app.TypesIndex(rr, req, httprouter.Params{})

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check the response body is what we expect.
	if strings.Contains(rr.Body.String(), "uva-ann2") == false {
		t.Errorf("Unexpected response: got [%s]. Does not include [%s]", rr.Body.String(), "uva-ann2")
	}
}
