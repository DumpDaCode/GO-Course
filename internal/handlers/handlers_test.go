package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
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
		{"Route does not exist", "/hello/world", "GET", http.StatusNotFound},
		{"Show login Page", "/user/login", "GET", http.StatusOK},
		{"Dashboard page", "/admin/dashboard", "GET", http.StatusOK},
		{"All reservations", "/admin/reservations-all", "GET", http.StatusOK},
		{"New reservations", "/admin/reservations-new", "GET", http.StatusOK},
		{"Show reservations", "/admin/reservations/new/1/show", "GET", http.StatusOK},
		{"Invalid Reservation id for Show reservations", "/admin/reservations/new/3/show", "GET", http.StatusInternalServerError},
		{"Invalid Reservation id type for Show reservations", "/admin/reservations/new/as/show", "GET", http.StatusInternalServerError},
		{"Reservation Calendar", "/admin/reservations-calendar?y=2050&m=05", "GET", http.StatusOK},
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
			want:    http.StatusOK,
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

// Doing this because we need session
func TestRepository_Logout(t *testing.T) {
	var testCases = []struct {
		name string
		want int
	}{
		{
			name: "Valid Case",
			want: http.StatusSeeOther,
		},
	}

	for _, testCase := range testCases {
		req, _ := http.NewRequest("GET", "/user/logout", nil)
		ctx := getCtx(req)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(Repo.Logout)
		handler.ServeHTTP(rr, req)
		if rr.Code != testCase.want {
			t.Errorf("Logout handler returned wrong response code for (%s): got %d, wanted %d", testCase.name, rr.Code, testCase.want)
		}
	}
}

func TestRepository_PostShowLogin(t *testing.T) {
	var testCases = []struct {
		name     string
		email    string
		want     int
		html     string
		location string
	}{
		{
			name:     "valid credentials",
			email:    "rajiv@mkcl.org",
			want:     http.StatusSeeOther,
			html:     "",
			location: "",
		},
		{
			name:     "Invalid credentials",
			email:    "jack@mkcl.org",
			want:     http.StatusSeeOther,
			html:     "",
			location: "/user/login",
		},
		{
			name:     "Invalid email",
			email:    "jack",
			want:     http.StatusOK,
			html:     "action=\"/user/login\"",
			location: "",
		},
	}

	for _, testCase := range testCases {
		postedData := url.Values{}
		postedData.Add("email", testCase.email)
		postedData.Add("password", "rajiv")
		req, _ := http.NewRequest("POST", "/user/login", strings.NewReader(postedData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		ctx := getCtx(req)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(Repo.PostShowLogin)
		handler.ServeHTTP(rr, req)
		if rr.Code != testCase.want {
			t.Errorf("ShowLogin handler returned wrong response code for (%s): got %d, wanted %d", testCase.name, rr.Code, testCase.want)
		} else if testCase.location != "" {
			actualLoc, _ := rr.Result().Location()
			if actualLoc.String() != testCase.location {
				t.Errorf("Redirected on wrong path for(%s): got %s, wanted %s", testCase.name, actualLoc.String(), testCase.location)
			}
		} else if testCase.html != "" {
			html := rr.Body.String()
			if !strings.Contains(html, testCase.html) {
				t.Log(strings.Contains(html, testCase.html))
				t.Errorf("Wrong HTML rendered for(%s): expected to find %s", testCase.name, testCase.html)
			}
		}
	}
}

func TestRepository_AdminPostReservationsCalendar(t *testing.T) {
	var testCases = []struct {
		name        string
		postData    url.Values
		sessionData map[string]int
		want        int
	}{
		{
			name: "valid case",
			postData: url.Values{
				"y":                         {"2050"},
				"m":                         {"01"},
				"add_block_1_2050-01-01":    {"1"},
				"remove_block_1_2050-01-04": {"1"},
				"remove_block_1_2050-01-06": {"2"},
			},
			sessionData: map[string]int{},
			want:        http.StatusSeeOther,
		},
		{
			name: "Adding a block ",
			postData: url.Values{
				"y":                         {"2050"},
				"m":                         {"01"},
				"add_block_1_2050-01-01":    {"1"},
				"remove_block_1_2050-01-04": {"1"},
				"remove_block_1_2050-01-06": {"2"},
			},
			sessionData: map[string]int{
				"2050-01-01": 1,
			},
			want: http.StatusSeeOther,
		},
		{
			name: "Updating a block ",
			postData: url.Values{
				"y":                         {"2050"},
				"m":                         {"01"},
				"add_block_1_2050-01-01":    {"1"},
				"remove_block_1_2050-01-04": {"1"},
				"remove_block_1_2050-01-06": {"2"},
			},
			sessionData: map[string]int{
				"2050-01-04": 1,
			},
			want: http.StatusSeeOther,
		},
		{
			name: "Not existing room",
			postData: url.Values{
				"y":                         {"2050"},
				"m":                         {"01"},
				"add_block_3_2050-01-01":    {"1"},
				"remove_block_3_2050-01-04": {"1"},
				"remove_block_3_2050-01-06": {"2"},
			},
			sessionData: map[string]int{
				"2050-01-01": 3,
			},
			want: http.StatusInternalServerError,
		},
	}

	for _, testCase := range testCases {
		req, _ := http.NewRequest("POST", "/admin/reservations-calendar", strings.NewReader(testCase.postData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		ctx := getCtx(req)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(Repo.AdminPostReservationsCalendar)
		session.Put(req.Context(), "block_map_1", testCase.sessionData)
		session.Put(req.Context(), "block_map_2", testCase.sessionData)
		handler.ServeHTTP(rr, req)

		if rr.Code != testCase.want {
			t.Errorf("AdminReservationsCalendar handler errored for (%s): got %d, want %d", testCase.name, rr.Code, testCase.want)
		}
	}
}

func TestRepository_AdminPostShowReservation(t *testing.T) {
	var testCases = []struct {
		name     string
		url      string
		postData url.Values
		want     int
	}{
		{
			name: "Valid case for all",
			url:  "/admin/reservations/all/1",
			postData: url.Values{
				"first_name": {"rajiv"},
				"last_name":  {"singh"},
				"email":      {"rajiv@gmail.org"},
				"phone":      {"8989898998"},
			},
			want: http.StatusSeeOther,
		},
		{
			name: "Valid case for new",
			url:  "/admin/reservations/new/1",
			postData: url.Values{
				"first_name": {"rajiv"},
				"last_name":  {"singh"},
				"email":      {"rajiv@gmail.org"},
				"phone":      {"8989898998"},
			},
			want: http.StatusSeeOther,
		},
		{
			name: "Valid case for cal",
			url:  "/admin/reservations/cal/1",
			postData: url.Values{
				"first_name": {"rajiv"},
				"last_name":  {"singh"},
				"email":      {"rajiv@gmail.org"},
				"phone":      {"8989898998"},
				"year":       {"2050"},
				"month":      {"01"},
			},
			want: http.StatusSeeOther,
		},
		{
			name: "Invalid id type",
			url:  "/admin/reservations/cal/as",
			postData: url.Values{
				"first_name": {"rajiv"},
				"last_name":  {"singh"},
				"email":      {"rajiv@gmail.org"},
				"phone":      {"8989898998"},
				"year":       {"2050"},
				"month":      {"01"},
			},
			want: http.StatusInternalServerError,
		},
		{
			name: "Invalid room",
			url:  "/admin/reservations/cal/3",
			postData: url.Values{
				"first_name": {"rajiv"},
				"last_name":  {"singh"},
				"email":      {"rajiv@gmail.org"},
				"phone":      {"8989898998"},
				"year":       {"2050"},
				"month":      {"01"},
			},
			want: http.StatusInternalServerError,
		},
		{
			name:     "Empty postData",
			url:      "/admin/reservations/cal/1",
			postData: nil,
			want:     http.StatusInternalServerError,
		},
		{
			name: "Invalid reservation",
			url:  "/admin/reservations/cal/1",
			postData: url.Values{
				"first_name": {"singh"},
				"last_name":  {"singh"},
				"email":      {"rajiv@gmail.org"},
				"phone":      {"8989898998"},
				"year":       {"2050"},
				"month":      {"01"},
			},
			want: http.StatusInternalServerError,
		},
	}

	for _, testCase := range testCases {
		var req *http.Request
		if testCase.postData == nil {
			req, _ = http.NewRequest("POST", testCase.url, nil)
		} else {
			req, _ = http.NewRequest("POST", testCase.url, strings.NewReader(testCase.postData.Encode()))
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		ctx := getCtx(req)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(Repo.AdminPostShowReservation)
		handler.ServeHTTP(rr, req)

		if rr.Code != testCase.want {
			t.Errorf("AdminPostShowReservation handler returned wrong status code for (%s): got %d, want %d", testCase.name, rr.Code, testCase.want)
		}
	}
}

func TestRepository_AdminProcessReservation(t *testing.T) {
	var testCases = []struct {
		name string
		url  string
		want int
	}{
		{
			name: "Valid Case for cal route",
			url:  "/admin/process-reservation/cal/1/do?y=2050&m=01",
			want: http.StatusSeeOther,
		},
		{
			name: "Valid Case for route other than cal",
			url:  "/admin/process-reservation/all/1/do",
			want: http.StatusSeeOther,
		},
		{
			name: "Invalid reservation id",
			url:  "/admin/process-reservation/all/3/do",
			want: http.StatusInternalServerError,
		},
		{
			name: "Invalid id type",
			url:  "/admin/process-reservation/all/as/do",
			want: http.StatusInternalServerError,
		},
	}

	for _, testCase := range testCases {
		var req *http.Request
		req, _ = http.NewRequest("POST", testCase.url, nil)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		ctx := getCtx(req)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(Repo.AdminProcessReservation)
		handler.ServeHTTP(rr, req)
		if rr.Code != testCase.want {
			t.Errorf("AdminProcessReservation handler returned wrong status code for (%s): got %d, want %d", testCase.name, rr.Code, testCase.want)
		}
	}
}

func TestRepository_AdminDeleteReservation(t *testing.T) {
	var testCases = []struct {
		name string
		url  string
		want int
	}{
		{
			name: "Valid Case for cal route",
			url:  "/admin/delete-reservation/cal/1/do?y=2050&m=01",
			want: http.StatusSeeOther,
		},
		{
			name: "Valid Case for route other than cal",
			url:  "/admin/delete-reservation/all/1/do",
			want: http.StatusSeeOther,
		},
		{
			name: "Invalid reservation id",
			url:  "/admin/delete-reservation/all/3/do",
			want: http.StatusInternalServerError,
		},
		{
			name: "Invalid id type",
			url:  "/admin/delete-reservation/all/as/do",
			want: http.StatusInternalServerError,
		},
	}

	for _, testCase := range testCases {
		var req *http.Request
		req, _ = http.NewRequest("POST", testCase.url, nil)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		ctx := getCtx(req)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(Repo.AdminDeleteReservation)
		handler.ServeHTTP(rr, req)
		if rr.Code != testCase.want {
			t.Errorf("AdminDeleteReservation handler returned wrong status code for (%s): got %d, want %d", testCase.name, rr.Code, testCase.want)
		}
	}
}
