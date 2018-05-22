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

func TestBadLookup(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Stub DB connection failed: %s", err)
	}
	defer mockDB.Close()
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	app := ApolloHandler{Version: "MOCK", DB: &models.DB{sqlxDB}}

	mock.ExpectQuery("SELECT np.pid").WillReturnError(errors.New("Not Found"))

	req, _ := http.NewRequest("GET", "/api/legacy/lookup", nil)
	rr := httptest.NewRecorder()
	app.ExternalPIDLookup(rr, req, httprouter.Params{httprouter.Param{"pid", "uva-lib:bad"}})

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusNotFound {
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
	app := ApolloHandler{Version: "MOCK", DB: &models.DB{sqlxDB}}

	rows := sqlmock.NewRows([]string{"pid"}).AddRow("uva-an1")
	mock.ExpectQuery("SELECT np.pid").WillReturnRows(rows) // Query is CASE SENSITIVE

	req, _ := http.NewRequest("GET", "/api/legacy/lookup", nil)
	rr := httptest.NewRecorder()
	app.ExternalPIDLookup(rr, req, httprouter.Params{httprouter.Param{"pid", "uva-lib:100"}})

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check the response body is what we expect.
	if strings.Contains(rr.Body.String(), "uva-an1") == false {
		t.Errorf("Unexpected response: got [%s]. Does not include [%s]", rr.Body.String(), "uva-an1")
	}
}
