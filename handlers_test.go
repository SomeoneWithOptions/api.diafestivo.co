package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/SomeoneWithOptions/api.diafestivo.co/holiday"
)

func TestRouteContracts(t *testing.T) {
	restore := holiday.SetNowFuncForTest(func() time.Time {
		return time.Date(2025, 6, 10, 12, 0, 0, 0, time.FixedZone("UTC-5", -5*60*60))
	})
	defer restore()

	mux := newServeMux()

	t.Run("GET /all", func(t *testing.T) {
		rec := performRequest(mux, http.MethodGet, "/all")
		assertStatus(t, rec, http.StatusOK)
		assertCORS(t, rec)
		assertContentType(t, rec, "application/json")

		var holidays []struct {
			Date string `json:"date"`
			Name string `json:"name"`
		}
		decodeJSON(t, rec, &holidays)
		if len(holidays) == 0 {
			t.Fatal("expected holidays")
		}
		if holidays[0].Name != "Año Nuevo" || holidays[0].Date != "2025-01-01T00:00:00Z" {
			t.Fatalf("first holiday = %+v, want Año Nuevo on 2025-01-01", holidays[0])
		}
	})

	t.Run("GET /make?year=2027", func(t *testing.T) {
		rec := performRequest(mux, http.MethodGet, "/make?year=2027")
		assertStatus(t, rec, http.StatusOK)
		assertCORS(t, rec)
		assertContentType(t, rec, "application/json")

		var holidays []struct {
			Date string `json:"date"`
			Name string `json:"name"`
		}
		decodeJSON(t, rec, &holidays)
		if holidays[0].Name != "Año Nuevo" || holidays[0].Date != "2027-01-01T00:00:00Z" {
			t.Fatalf("first holiday = %+v, want Año Nuevo on 2027-01-01", holidays[0])
		}
		if !containsHoliday(holidays, "el Día de Nuestra Señora del Rosario de Chiquinquirá", "2027-07-12T00:00:00Z") {
			t.Fatal("expected Chiquinquirá holiday in 2027")
		}
	})

	t.Run("GET /is/2025-06-11", func(t *testing.T) {
		rec := performRequest(mux, http.MethodGet, "/is/2025-06-11")
		assertStatus(t, rec, http.StatusOK)
		assertCORS(t, rec)
		assertContentType(t, rec, "application/json")
		var body map[string]bool
		decodeJSON(t, rec, &body)
		if body["isHoliday"] {
			t.Fatal("2025-06-11 should not be holiday")
		}
	})

	t.Run("GET /is/2025-01-01", func(t *testing.T) {
		rec := performRequest(mux, http.MethodGet, "/is/2025-01-01")
		assertStatus(t, rec, http.StatusOK)
		assertCORS(t, rec)
		assertContentType(t, rec, "application/json")
		var body map[string]bool
		decodeJSON(t, rec, &body)
		if !body["isHoliday"] {
			t.Fatal("2025-01-01 should be holiday")
		}
	})

	t.Run("GET /is/bad", func(t *testing.T) {
		rec := performRequest(mux, http.MethodGet, "/is/bad")
		assertStatus(t, rec, http.StatusBadRequest)
		assertCORS(t, rec)
		if body := strings.TrimSpace(rec.Body.String()); body != "error parsing date" {
			t.Fatalf("body = %q, want error parsing date", body)
		}
	})

	t.Run("GET /make?year=bad", func(t *testing.T) {
		rec := performRequest(mux, http.MethodGet, "/make?year=bad")
		assertStatus(t, rec, http.StatusBadRequest)
		assertCORS(t, rec)
		if body := strings.TrimSpace(rec.Body.String()); body != "error parsing year" {
			t.Fatalf("body = %q, want error parsing year", body)
		}
	})

	t.Run("GET /next", func(t *testing.T) {
		rec := performRequest(mux, http.MethodGet, "/next")
		assertStatus(t, rec, http.StatusOK)
		assertCORS(t, rec)
		assertContentType(t, rec, "application/json")

		var body struct {
			Name      string `json:"name"`
			Date      string `json:"date"`
			IsToday   bool   `json:"isToday"`
			DaysUntil int    `json:"daysUntil"`
		}
		decodeJSON(t, rec, &body)
		if body.Name != "Corpus Christi" || body.Date != "2025-06-23T00:00:00Z" || body.IsToday || body.DaysUntil != 13 {
			t.Fatalf("/next body = %+v, want Corpus Christi in 13 days", body)
		}
	})

	t.Run("GET /template", func(t *testing.T) {
		rec := performRequest(mux, http.MethodGet, "/template")
		assertStatus(t, rec, http.StatusOK)
		assertCORS(t, rec)
		assertContentType(t, rec, "text/html")
		body := rec.Body.String()
		if !strings.Contains(body, "NO :(") || !strings.Contains(body, "Corpus Christi") {
			t.Fatalf("/template body missing expected current/next holiday text: %q", body)
		}
	})

	t.Run("GET /left", func(t *testing.T) {
		rec := performRequest(mux, http.MethodGet, "/left")
		assertStatus(t, rec, http.StatusOK)
		assertCORS(t, rec)
		assertContentType(t, rec, "text/html")
		body := rec.Body.String()
		if !strings.Contains(body, "Corpus Christi") || !strings.Contains(body, "en 13 Días") {
			t.Fatalf("/left body missing expected remaining holiday: %q", body)
		}
	})

	t.Run("invalid route", func(t *testing.T) {
		rec := performRequest(mux, http.MethodGet, "/not-a-route")
		assertStatus(t, rec, http.StatusBadRequest)
		assertCORS(t, rec)
		assertContentType(t, rec, "application/json")

		var body struct {
			Status      int      `json:"status"`
			Message     string   `json:"message"`
			ValidRoutes []string `json:"valid_routes"`
		}
		decodeJSON(t, rec, &body)
		if body.Status != http.StatusBadRequest || body.Message != "Please Use Valid Routes:" || len(body.ValidRoutes) != 4 {
			t.Fatalf("invalid-route body = %+v", body)
		}
	})

	t.Run("GET /healthz", func(t *testing.T) {
		rec := performRequest(mux, http.MethodGet, "/healthz")
		assertStatus(t, rec, http.StatusOK)
		assertCORS(t, rec)
		if body := strings.TrimSpace(rec.Body.String()); body != "ok" {
			t.Fatalf("body = %q, want ok", body)
		}
	})
}

func performRequest(handler http.Handler, method, target string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, target, nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

func assertStatus(t *testing.T, rec *httptest.ResponseRecorder, want int) {
	t.Helper()
	if rec.Code != want {
		t.Fatalf("status = %d, want %d; body=%q", rec.Code, want, rec.Body.String())
	}
}

func assertCORS(t *testing.T, rec *httptest.ResponseRecorder) {
	t.Helper()
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Fatalf("CORS header = %q, want *", got)
	}
}

func assertContentType(t *testing.T, rec *httptest.ResponseRecorder, want string) {
	t.Helper()
	if got := rec.Header().Get("Content-Type"); got != want {
		t.Fatalf("Content-Type = %q, want %q", got, want)
	}
}

func decodeJSON(t *testing.T, rec *httptest.ResponseRecorder, target any) {
	t.Helper()
	if err := json.Unmarshal(rec.Body.Bytes(), target); err != nil {
		t.Fatalf("failed to decode JSON %q: %v", rec.Body.String(), err)
	}
}

func containsHoliday(holidays []struct {
	Date string `json:"date"`
	Name string `json:"name"`
}, name, date string) bool {
	for _, holiday := range holidays {
		if holiday.Name == name && holiday.Date == date {
			return true
		}
	}
	return false
}
