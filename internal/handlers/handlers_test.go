package handlers

import (
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-course/bookings/internal/models"
)

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
	// {"Post Search Availability", "/search-availability", "POST", []postData{
	// 	{key: "start", value: "2020-01-01"},
	// 	{key: "end", value: "2020-01-01"},
	// }, http.StatusOK},
	// {"Post Search Availability JSON", "/search-availability-json", "POST", []postData{
	// 	{key: "start", value: "2020-01-01"},
	// 	{key: "end", value: "2020-01-01"},
	// }, http.StatusOK},
	// {"Post make reservation", "/make-reservation", "POST", []postData{
	// 	{key: "first_name", value: "Rajiv"},
	// 	{key: "last_name", value: "singh"},
	// 	{key: "email", value: "me@here.com"},
	// 	{key: "phone", value: "555-555-5555"},
	// }, http.StatusOK},
}

func TestHandlers(t *testing.T) {
	routes := getRoutes()
	ts := httptest.NewTLSServer(routes)
	defer ts.Close()

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
	reservation := models.Reservation{
		RoomID: 1,
		Room: models.Room{
			ID:       1,
			RoomName: "General's Quarters",
		},
	}

	req, _ := http.NewRequest("GET", "/make-reservation", nil)
	ctx := getCtx(req)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	session.Put(ctx, "reservation", reservation)
	handler := http.HandlerFunc(Repo.Reservation)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Reservation handler returned wrong response code: got %d, wanted %d", rr.Code, http.StatusOK)
	}

	// test case wher reservation is not in sesion (reset everything)
	req, _ = http.NewRequest("GET", "/make-reservation", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)

	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("Reservation handler returned wrong response code: got %d, wanted %d", rr.Code, http.StatusTemporaryRedirect)
	}

	// test with non existent room
	req, _ = http.NewRequest("GET", "/make-reservation", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)

	rr = httptest.NewRecorder()
	reservation.RoomID = 100
	session.Put(ctx, "reservation", reservation)
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("Reservation handler returned wrong response code: got %d, wanted %d", rr.Code, http.StatusTemporaryRedirect)
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

func getCtx(req *http.Request) context.Context {
	ctx, err := session.Load(req.Context(), req.Header.Get("X-Session"))
	if err != nil {
		log.Println(err)
	}
	return ctx
}
