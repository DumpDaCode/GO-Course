package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-course/bookings/internal/models"
)

func TestGetHandlers(t *testing.T) {
	routes := getRoutes()
	ts := httptest.NewTLSServer(routes)
	defer ts.Close()

	var theTests = []struct {
		name               string
		url                string
		method             string
		expectedStatusCode int
	}{
		{"Home", "/", "GET", http.StatusOK},
		{"about", "/about", "GET", http.StatusOK},
		{"Generals Quarters", "/generals-quarters", "GET", http.StatusOK},
		{"Major's Suite", "/majors-suite", "GET", http.StatusOK},
		{"Search Availability", "/search-availability", "GET", http.StatusOK},
		{"Contact", "/contact", "GET", http.StatusOK},
	}

	for _, e := range theTests {
		resp, err := ts.Client().Get(ts.URL + e.url)
		if err != nil {
			t.Log(e)
			t.Fatal(err)
		}
		if resp.StatusCode != e.expectedStatusCode {
			t.Errorf("for %s expected %d but got %d", e.name, e.expectedStatusCode, resp.StatusCode)
		}
	}
}

func TestRepository_Reservation(t *testing.T) {
	var testCases = []struct {
		name    string
		payload *models.Reservation
		want    int
	}{
		{
			name: "Valid case",
			payload: &models.Reservation{
				RoomID: 1,
				Room: models.Room{
					ID:       1,
					RoomName: "General's Quarters",
				},
			},
			want: http.StatusOK,
		},
		{
			name:    "Reservation not in session",
			payload: nil,
			want:    http.StatusTemporaryRedirect,
		},
		{
			name: "Non existent room",
			payload: &models.Reservation{
				RoomID: 100,
				Room: models.Room{
					ID:       1,
					RoomName: "General's Quarters",
				},
			},
			want: http.StatusTemporaryRedirect,
		},
	}
	for _, testCase := range testCases {
		req, _ := http.NewRequest("GET", "/make-reservation", nil)
		ctx := getCtx(req)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		if testCase.payload != nil {
			session.Put(ctx, "reservation", *testCase.payload)
		}
		handler := http.HandlerFunc(Repo.Reservation)
		handler.ServeHTTP(rr, req)

		if rr.Code != testCase.want {
			t.Errorf("Reservation handler returned wrong response code for (%s): got %d, wanted %d", testCase.name, rr.Code, testCase.want)
		}
	}
}

func TestRepository_PostReservation(t *testing.T) {
	var testCases = []struct {
		name    string
		reqBody *strings.Reader
		want    int
	}{
		{
			name:    "Valid case",
			reqBody: strings.NewReader("start_date=2050-01-01&end_date=2050-01-02&first_name=John&last_name=Smith&email=rajiv@mkcl.org&phone=123456789&room_id=1"),
			want:    http.StatusSeeOther,
		},
		{
			name:    "Body missing",
			reqBody: nil,
			want:    http.StatusTemporaryRedirect,
		},
		{
			name:    "Start Date absent",
			reqBody: strings.NewReader("end_date=2050-01-02&first_name=John&last_name=Smith&email=rajiv@mkcl.org&phone=123456789&room_id=1"),
			want:    http.StatusTemporaryRedirect,
		},
		{
			name:    "End Date absent",
			reqBody: strings.NewReader("start_date=2050-01-01&first_name=John&last_name=Smith&email=rajiv@mkcl.org&phone=123456789&room_id=1"),
			want:    http.StatusTemporaryRedirect,
		},
		{
			name:    "Room Id absent",
			reqBody: strings.NewReader("start_date=2050-01-01&end_date=2050-01-02&first_name=John&last_name=Smith&email=rajiv@mkcl.org&phone=123456789"),
			want:    http.StatusTemporaryRedirect,
		},
		{
			name:    "Form validation failed",
			reqBody: strings.NewReader("start_date=2050-01-01&end_date=2050-01-02&first_name=J&last_name=Smith&email=rajiv@mkcl.org&phone=123456789&room_id=1"),
			want:    http.StatusSeeOther,
		},
		{
			name:    "Unable to add reservation",
			reqBody: strings.NewReader("start_date=2050-01-01&end_date=2050-01-02&first_name=John&last_name=Smith&email=rajiv@mkcl.org&phone=123456789&room_id=2"),
			want:    http.StatusTemporaryRedirect,
		},
		{
			name:    "Unable to add restriction",
			reqBody: strings.NewReader("start_date=2050-01-01&end_date=2050-01-02&first_name=John&last_name=Smith&email=rajiv@mkcl.org&phone=123456789&room_id=100"),
			want:    http.StatusTemporaryRedirect,
		},
	}

	for _, testCase := range testCases {
		var req *http.Request
		if testCase.reqBody == nil {
			req, _ = http.NewRequest("POST", "/make-reservation", nil)
		} else {
			req, _ = http.NewRequest("POST", "/make-reservation", testCase.reqBody)
		}
		ctx := getCtx(req)
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(Repo.PostReservation)
		handler.ServeHTTP(rr, req)

		if rr.Code != testCase.want {
			t.Errorf("PostReservation handler returned wrong response code for test case (%s) : got %d, wanted %d", testCase.name, rr.Code, testCase.want)
		}
	}
}

func TestRepository_AvailabilityJSON(t *testing.T) {
	var testCases = []struct {
		name    string
		reqBody *strings.Reader
		want    bool
	}{
		{
			name:    "Valid case",
			reqBody: strings.NewReader("start=2050-01-01&end=2050-01-02&room_id=1"),
			want:    true,
		},
		{
			name:    "No body",
			reqBody: nil,
			want:    false,
		},
		{
			name:    "Invalid room",
			reqBody: strings.NewReader("end=2050-01-02&room_id=100"),
			want:    false,
		},
	}
	for _, testCase := range testCases {
		var req *http.Request
		if testCase.reqBody == nil {
			req, _ = http.NewRequest("POST", "/search-availability-json", nil)
		} else {
			req, _ = http.NewRequest("POST", "/search-availability-json", testCase.reqBody)
		}

		ctx := getCtx(req)
		req = req.WithContext(ctx)

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		rr := httptest.NewRecorder()

		handler := http.HandlerFunc(Repo.AvailabilityJSON)

		handler.ServeHTTP(rr, req)

		var j jsonResponse
		err := json.Unmarshal(rr.Body.Bytes(), &j)
		if err != nil {
			t.Error("failed to parse json")
		}
		if j.OK != testCase.want {
			t.Errorf("AvailabilityJSON handler returned wrong response code for test case (%s) : got %v, wanted %v", testCase.name, j.OK, testCase.want)
		}
	}
}

func getCtx(req *http.Request) context.Context {
	ctx, err := session.Load(req.Context(), req.Header.Get("X-Session"))
	if err != nil {
		log.Println(err)
	}
	return ctx
}

func TestRepository_PostAvailability(t *testing.T) {
	var testCases = []struct {
		name    string
		reqBody *strings.Reader
		want    int
	}{
		{
			name:    "Valid case",
			reqBody: strings.NewReader("start=2050-01-01&end=2050-01-02"),
			want:    http.StatusOK,
		},
		{
			name:    "No body",
			reqBody: nil,
			want:    http.StatusSeeOther,
		},
		{
			name:    "No Start Date",
			reqBody: strings.NewReader("end=2050-01-02"),
			want:    http.StatusSeeOther,
		},
		{
			name:    "No End Date",
			reqBody: strings.NewReader("start=2050-01-01"),
			want:    http.StatusSeeOther,
		},
		{
			name:    "Rooms not available",
			reqBody: strings.NewReader("start=2022-08-27&end=2050-01-02"),
			want:    http.StatusSeeOther,
		},
	}
	for _, testCase := range testCases {
		var req *http.Request
		if testCase.reqBody == nil {
			req, _ = http.NewRequest("POST", "/search-availability", nil)
		} else {
			req, _ = http.NewRequest("POST", "/search-availability", testCase.reqBody)
		}

		ctx := getCtx(req)
		req = req.WithContext(ctx)

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		rr := httptest.NewRecorder()

		handler := http.HandlerFunc(Repo.PostAvailability)

		handler.ServeHTTP(rr, req)

		if rr.Code != testCase.want {
			t.Errorf("AvailabilityJSON handler returned wrong response code for test case (%s) : got %v, wanted %v", testCase.name, rr.Code, testCase.want)
		}
	}
}

func TestRepository_ReservationSummary(t *testing.T) {
	var testCases = []struct {
		name    string
		payload *models.Reservation
		want    int
	}{
		{
			name: "Valid case",
			payload: &models.Reservation{
				RoomID:    1,
				StartDate: time.Now(),
				EndDate:   time.Now(),
			},
			want: http.StatusOK,
		},
		{
			name:    "Reservation not in session",
			payload: nil,
			want:    http.StatusTemporaryRedirect,
		},
	}
	for _, testCase := range testCases {
		req, _ := http.NewRequest("GET", "/make-reservation", nil)
		ctx := getCtx(req)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		if testCase.payload != nil {
			session.Put(ctx, "reservation", *testCase.payload)
		}
		handler := http.HandlerFunc(Repo.ReservationSummary)
		handler.ServeHTTP(rr, req)

		if rr.Code != testCase.want {
			t.Errorf("ReservationSummary handler returned wrong response code for (%s): got %d, wanted %d", testCase.name, rr.Code, testCase.want)
		}
	}
}

func TestRepository_ChooseRoom(t *testing.T) {
	var testCases = []struct {
		name    string
		reqURI  string
		payload *models.Reservation
		want    int
	}{
		{
			name:    "Valid case",
			reqURI:  "/choose-room/1",
			payload: &models.Reservation{},
			want:    http.StatusSeeOther,
		},
		{
			name:    "Reservation is absent from session",
			reqURI:  "/choose-room/1",
			payload: nil,
			want:    http.StatusTemporaryRedirect,
		},
		{
			name:    "Invalid Room ID",
			reqURI:  "/choose-room/as",
			payload: &models.Reservation{},
			want:    http.StatusTemporaryRedirect,
		},
	}
	for _, testCase := range testCases {
		req, _ := http.NewRequest("GET", "", nil)
		ctx := getCtx(req)
		req = req.WithContext(ctx)
		req.RequestURI = testCase.reqURI
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(Repo.ChooseRoom)
		if testCase.payload != nil {
			session.Put(ctx, "reservation", *testCase.payload)
		}
		handler.ServeHTTP(rr, req)
		if rr.Code != testCase.want {
			t.Errorf("ChooseRoom handler returned wrong response code for (%s): got %d, wanted %d", testCase.name, rr.Code, testCase.want)
		}
	}
}

func TestRepository_BookRoom(t *testing.T) {
	var testCases = []struct {
		name   string
		reqURL string
		want   int
	}{
		{
			name:   "Valid case",
			reqURL: "/book-room?id=1&s=2050-01-02&e=2050-01-02",
			want:   http.StatusSeeOther,
		},
		{
			name:   "Invalid Id",
			reqURL: "/book-room?id=asd&s=2050-01-02&e=2050-01-02",
			want:   http.StatusTemporaryRedirect,
		},
		{
			name:   "No start date",
			reqURL: "/book-room?id=1&e=2050-01-02",
			want:   http.StatusTemporaryRedirect,
		},
		{
			name:   "No end date",
			reqURL: "/book-room?id=1&s=2050-01-02",
			want:   http.StatusTemporaryRedirect,
		},
		{
			name:   "For unavailable room",
			reqURL: "/book-room?id=3&s=2050-01-02&e=2050-01-02",
			want:   http.StatusTemporaryRedirect,
		},
	}

	for _, testCase := range testCases {
		req, _ := http.NewRequest("GET", testCase.reqURL, nil)
		ctx := getCtx(req)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(Repo.BookRoom)
		handler.ServeHTTP(rr, req)

		if rr.Code != testCase.want {
			t.Errorf("BookRoom handler returned wrong response code for (%s): got %d, wanted %d", testCase.name, rr.Code, testCase.want)
		}
	}
}
