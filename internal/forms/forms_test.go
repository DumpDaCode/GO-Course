package forms

import (
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestForm_Valid(t *testing.T) {
	r := httptest.NewRequest("POST", "/whatever", nil)
	form := New(r.PostForm)
	isValid := form.Valid()
	if !isValid {
		t.Error("got invalid when should have been valid")
	}
}

func TestForm_Required(t *testing.T) {
	form := New(url.Values{})

	form.Required("a", "b", "c")
	if form.Valid() {
		t.Error("form shows valid when required field missing.")
	}

	postedData := url.Values{}
	postedData.Add("a", "a")
	postedData.Add("b", "b")
	postedData.Add("c", "c")

	form = New(postedData)

	form.Required("a", "b", "c")
	if !form.Valid() {
		t.Error("form does not have required fields when it does.")
	}

}

func TestForm_Has(t *testing.T) {
	form := New(url.Values{})
	isValid := form.Has("a")
	if isValid {
		t.Error("form shows valid when required field missing.")
	}

	postedData := url.Values{}
	postedData.Add("a", "a")
	form = New(postedData)
	isValid = form.Has("a")
	if !isValid {
		t.Error("form does not have required fields when it does.")
	}
}

func TestForm_MinLength(t *testing.T) {
	form := New(url.Values{})
	isValid := form.MinLength("a", 3)
	if isValid {
		t.Error("form shows valid when minlength criteria is invalid. (EMPTY)")
	}

	postedData := url.Values{}
	postedData.Add("a", "abcd")
	postedData.Add("b", "ab")
	form = New(postedData)
	isValid = form.MinLength("a", 3)
	if !isValid {
		t.Error("form does not show valid even when minlength criteria is valid..")
	}

	isError := form.Errors.Get("a")
	if isError != "" {
		t.Error("should not have an error but got one")
	}

	isValid = form.MinLength("b", 3)
	if isValid {
		t.Error("form shows valid when minlength criteria is invalid. (NOT EMPTY)")
	}

	isError = form.Errors.Get("b")
	if isError == "" {
		t.Error("should have an error but didn't get one")
	}

}

func TestForm_IsEmail(t *testing.T) {
	form := New(url.Values{})
	form.IsEmail("a")
	if form.Valid() {
		t.Error("form shows valid when email criteria is invalid. (EMPTY)")
	}

	postedData := url.Values{}
	postedData.Add("a", "rajiv@mkcl.org")
	postedData.Add("b", "rajiv@.org")
	form = New(postedData)
	form.IsEmail("a")
	if !form.Valid() {
		t.Error("form does not show valid even when email criteria is valid.")
	}

	form.IsEmail("b")
	if form.Valid() {
		t.Error("form shows valid when email criteria is invalid. (NOT EMPTY)")
	}
}
