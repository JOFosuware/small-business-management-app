package forms

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestForm_Valid(t *testing.T) {
	r := httptest.NewRequest("POST", "/some-url", nil)
	form := New(r.PostForm)

	isValid := form.Valid()
	if !isValid {
		t.Error("got invalid when should have been valid")
	}
}

func TestForm_Required(t *testing.T) {
	r := httptest.NewRequest("POST", "/whatever", nil)
	form := New(r.PostForm)

	form.Required("a", "b", "c")
	if form.Valid() {
		t.Error("form shows valid when required fields missing")
	}

	postedData := url.Values{}
	postedData.Add("a", "a")
	postedData.Add("b", "a")
	postedData.Add("c", "a")

	r, _ = http.NewRequest("POST", "/whatever", nil)

	r.PostForm = postedData
	form = New(r.PostForm)
	form.Required("a", "b", "c")
	if !form.Valid() {
		t.Error("shows does not have required fields when it does")
	}
}

func TestForm_Has(t *testing.T) {
	r := httptest.NewRequest("POST", "/some-url", nil)
	form := New(r.PostForm)

	has := form.Has("whatever")

	if has {
		t.Error("form shows has field when it does not")
	}

	postedForm := url.Values{}
	postedForm.Add("a", "a")
	form = New(postedForm)

	has = form.Has("a")
	if !has {
		t.Error("shows form does not have field when it should")
	}
}

func TestForm_MinLength(t *testing.T) {
	r := httptest.NewRequest("POST", "/some-url", nil)
	form := New(r.PostForm)

	form.MinLength("x", 8)
	if form.Valid() {
		t.Error("form shows min length for non-existent field")
	}

	isError := form.Errors.Get("x")
	if isError == "" {
		t.Error("should have an error, but did not get one")
	}

	postedForm := url.Values{}
	postedForm.Add("password", "123")
	form = New(postedForm)

	form.MinLength("password", 8)
	if form.Valid() {
		t.Error("show minlength 0f 8 met when data is shorter")
	}

	postedForm = url.Values{}
	postedForm.Add("password1", "Software@2020")
	form = New(postedForm)

	isError = form.Errors.Get("password1")
	if isError != "" {
		t.Error("should not have an error, but got one")
	}
}

func TestForm_IsEmail(t *testing.T) {
	postedForm := url.Values{}
	form := New(postedForm)

	form.IsEmail(form.Get("email"))
	if form.Valid() {
		t.Error("form shows valid email for non-existent field")
	}

	postedForm = url.Values{}
	postedForm.Add("email", "developer.com")
	form = New(postedForm)

	form.IsEmail("email")
	if form.Valid() {
		t.Error("form shows valid email for invalid email")
	}

	postedForm = url.Values{}
	postedForm.Add("email", "developer@deperidot.com")
	form = New(postedForm)

	form.IsEmail("email")
	if !form.Valid() {
		t.Error("form shows error for valid email")
	}
}
