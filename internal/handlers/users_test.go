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
	"github.com/uvalib/apollo/internal/models"
)

func TestUsersIndex(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Stub DB connection failed: %s", err)
	}
	defer mockDB.Close()
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	app := ApolloHandler{Version: "MOCK", DB: &models.DB{sqlxDB}}

	router := httprouter.New()
	router.GET("/api/users", app.UsersIndex)

	req, _ := http.NewRequest("GET", "/api/users", nil)
	rr := httptest.NewRecorder()

	rows := sqlmock.NewRows([]string{"id", "email"}).
		AddRow(1, "test1@virginia.edu").AddRow(2, "test2@virginia.edu")
	mock.ExpectQuery("select").WillReturnRows(rows)

	router.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check the response body is what we expect.
	expected := `[{"id": 1, "email": "test1@virginia.edu"},{"id": 2, "email": "test2@virginia.edu"}]`
	if strings.TrimSpace(rr.Body.String()) != expected {
		t.Errorf("Unexpected response: got [%s] want [%s]", rr.Body.String(), expected)
	}
}

func TestBadUserShow(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Stub DB connection failed: %s", err)
	}
	defer mockDB.Close()
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	app := ApolloHandler{Version: "MOCK", DB: &models.DB{sqlxDB}}
	mock.ExpectQuery("select").WillReturnError(errors.New("User not found"))

	req, _ := http.NewRequest("GET", "/api/users", nil)
	rr := httptest.NewRecorder()

	params := httprouter.Params{httprouter.Param{"id", "666"}}
	app.UsersShow(rr, req, params)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("Wrong status code: got %v want %v", status, http.StatusOK)
	}
}

func TestUserShow(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Stub DB connection failed: %s", err)
	}
	defer mockDB.Close()
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	app := ApolloHandler{Version: "MOCK", DB: &models.DB{sqlxDB}}

	rows := sqlmock.NewRows([]string{"id", "computing_id", "first_name", "last_name", "email"}).
		AddRow(1, "test1", "test", "user", "test1@virginia.edu")
	mock.ExpectQuery("SELECT").WillReturnRows(rows) // Query is CASE SENSITIVE

	req, _ := http.NewRequest("GET", "/api/users", nil)
	rr := httptest.NewRecorder()

	params := httprouter.Params{httprouter.Param{"id", "1"}}
	app.UsersShow(rr, req, params)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Wrong status code: got %v want %v", status, http.StatusOK)
	}
	// Check the response body is what we expect.
	expected := `{"id":1,"computingId":"test1","firstName":"test","lastName":"user","email":"test1@virginia.edu"}`
	if strings.TrimSpace(rr.Body.String()) != expected {
		t.Errorf("Unexpected response: got [%s] want [%s]", rr.Body.String(), expected)
	}
}
