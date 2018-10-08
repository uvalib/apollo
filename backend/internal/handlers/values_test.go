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

func TestGetValues(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Stub DB connection failed: %s", err)
	}
	defer mockDB.Close()
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	app := ApolloHandler{Version: "MOCK", DB: &models.DB{DB: sqlxDB}}

	rows := sqlmock.NewRows([]string{"pid", "value", "value_uri"}).
		AddRow("uva-acv1", "fake", "http://fake.com").AddRow("uva-acv2", "fake2", "http://fake2.com")
	mock.ExpectQuery("SELECT").WillReturnRows(rows) // Query is CASE SENSITIVE

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.GET("/api/values/:name", app.ValuesForName)

	req, _ := http.NewRequest("GET", "/api/values/fake", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check the response body is what we expect.
	if strings.Contains(rr.Body.String(), "uva-acv2") == false {
		t.Errorf("Unexpected response: got [%s]. Does not include [%s]", rr.Body.String(), "uva-acv2")
	}
}
