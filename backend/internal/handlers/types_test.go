package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/uvalib/apollo/backend/internal/models"
)

func TestTypesIndex(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Stub DB connection failed: %s", err)
	}
	defer mockDB.Close()
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	app := Apollo{Version: "MOCK", DB: &models.DB{DB: sqlxDB}}

	rows := sqlmock.NewRows([]string{"pid", "name"}).
		AddRow("uva-ann1", "collection").AddRow("uva-ann2", "title")
	mock.ExpectQuery("select").WillReturnRows(rows) // Query is CASE SENSITIVE

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.GET("/api/types", app.TypesIndex)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/types", nil)
	router.ServeHTTP(w, req)

	// Check the status code is what we expect.
	if status := w.Code; status != http.StatusOK {
		t.Errorf("Wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check the response body is what we expect.
	if strings.Contains(w.Body.String(), "uva-ann2") == false {
		t.Errorf("Unexpected response: got [%s]. Does not include [%s]", w.Body.String(), "uva-ann2")
	}
}
