package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/julienschmidt/httprouter"
	"github.com/uvalib/apollo/backend/internal/models"
)

func TestNoAuth(t *testing.T) {
	mockDB, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Stub DB connection failed: %s", err)
	}
	defer mockDB.Close()
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	app := ApolloHandler{Version: "MOCK", DB: &models.DB{sqlxDB}}

	router := httprouter.New()
	router.GET("/api/dummy", app.AuthHandler(dummyHandler))

	req, _ := http.NewRequest("GET", "/api/dummy", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusForbidden {
		t.Errorf("Wrong status code: got %v want %v", status, http.StatusForbidden)
	}
}

func TestGoodAuth(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Stub DB connection failed: %s", err)
	}
	defer mockDB.Close()
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	app := ApolloHandler{Version: "MOCK", DB: &models.DB{sqlxDB}, DevAuthUser: "lf6f"}

	rows := sqlmock.NewRows([]string{"id", "computing_id", "first_name", "last_name", "email"}).
		AddRow(1, "lf6f", "Lou", "Foster", "lf6f")
	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	router := httprouter.New()
	router.GET("/api/dummy", app.AuthHandler(dummyHandler))

	req, _ := http.NewRequest("GET", "/api/dummy", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Wrong status code: got %v want %v", status, http.StatusOK)
	}

	expected := `accessed`
	if strings.TrimSpace(rr.Body.String()) != expected {
		t.Errorf("Unexpected response: got [%s] want [%s]", rr.Body.String(), expected)
	}
}

func TestBadAuth(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Stub DB connection failed: %s", err)
	}
	defer mockDB.Close()
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	app := ApolloHandler{Version: "MOCK", DB: &models.DB{sqlxDB}, DevAuthUser: "BAD"}

	mock.ExpectQuery("SELECT").WillReturnError(errors.New("You are not authorized to access this site"))

	router := httprouter.New()
	router.GET("/api/dummy", app.AuthHandler(dummyHandler))

	req, _ := http.NewRequest("GET", "/api/dummy", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusForbidden {
		t.Errorf("Wrong status code: got %v want %v", status, http.StatusForbidden)
	}
}

func dummyHandler(rw http.ResponseWriter, req *http.Request, params httprouter.Params) {
	fmt.Fprintf(rw, "accessed")
}