package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/uvalib/apollo/backend/internal/models"
)

func TestUsersIndex(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Stub DB connection failed: %s", err)
	}
	defer mockDB.Close()
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	app := Apollo{Version: "MOCK", DB: &models.DB{DB: sqlxDB}}

	router := gin.Default()
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
	if strings.Contains(rr.Body.String(), "test2@virginia.edu") == false {
		t.Errorf("Unexpected response: got [%s]. Does not include [%s]", rr.Body.String(), "test2@virginia.edu")
	}
}

func TestBadUserShow(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Stub DB connection failed: %s", err)
	}
	defer mockDB.Close()
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	app := Apollo{Version: "MOCK", DB: &models.DB{DB: sqlxDB}}
	mock.ExpectQuery("select").WillReturnError(errors.New("User not found"))

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.GET("/api/users/:id", app.UsersShow)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/users/666", nil)
	router.ServeHTTP(w, req)

	// Check the status code is what we expect.
	if status := w.Code; status != http.StatusNotFound {
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
	app := Apollo{Version: "MOCK", DB: &models.DB{DB: sqlxDB}}

	rows := sqlmock.NewRows([]string{"id", "computing_id", "first_name", "last_name", "email"}).
		AddRow(1, "test1", "test", "user", "test1@virginia.edu")
	mock.ExpectQuery("SELECT").WillReturnRows(rows) // Query is CASE SENSITIVE

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.GET("/api/users/:id", app.UsersShow)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/users/1", nil)
	router.ServeHTTP(w, req)

	// Check the status code is what we expect.
	if status := w.Code; status != http.StatusOK {
		t.Errorf("Wrong status code: got %v want %v", status, http.StatusOK)
	}
	// Check the response body is what we expect.
	if strings.Contains(w.Body.String(), "test1@virginia.edu") == false {
		t.Errorf("Unexpected response: got [%s]. Does not include [%s]", w.Body.String(), "test1@virginia.edu")
	}
}
