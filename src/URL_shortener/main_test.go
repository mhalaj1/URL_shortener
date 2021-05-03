package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDetermineHandler(t *testing.T) {
	indexPageHandler := func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "indexPageHandler called")
	}
	redirectHandler := func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "redirectHandler called")
	}

	cases := []struct {
		target   string
		wantCode int
		wantBody string
	}{
		// indexPageHandler
		{"http://localhost:8080", 200, "indexPageHandler called"},
		{"http://localhost:8080/", 200, "indexPageHandler called"},
		{"http://localhost:8080/index.html", 200, "indexPageHandler called"},
		// redirectPageHandler
		{"http://localhost:8080/abcdef", 200, "redirectHandler called"},
	}

	for _, c := range cases {
		req := httptest.NewRequest("GET", c.target, nil)
		w := httptest.NewRecorder()
		DetermineHandler(w, req, indexPageHandler, redirectHandler)

		resp := w.Result()
		gotCode := resp.StatusCode
		gotBodyByte, _ := io.ReadAll(resp.Body)
		gotBody := string(gotBodyByte)

		if gotCode != c.wantCode {
			t.Errorf("Requested target %v, got code %v, want code %v", c.target, gotCode, c.wantCode)
		}
		if gotBody != c.wantBody {
			t.Errorf("Requested target %v, got body %v, want body %v", c.target, gotBody, c.wantBody)
		}
	}
}

func TestIndexPageHandler(t *testing.T) {
	serveFileFunc := func(w http.ResponseWriter, r *http.Request, name string) {
		fmt.Fprintf(w, "Index page served")
	}

	cases := []struct {
		target   string
		wantCode int
		wantBody string
	}{
		{"http://localhost:8080/", 200, "Index page served"},
	}

	for _, c := range cases {
		req := httptest.NewRequest("GET", c.target, nil)
		w := httptest.NewRecorder()
		IndexPageHandler(w, req, serveFileFunc)

		resp := w.Result()
		gotCode := resp.StatusCode
		gotBodyByte, _ := io.ReadAll(resp.Body)
		gotBody := string(gotBodyByte)

		if gotCode != c.wantCode {
			t.Errorf("Requested target %v, got code %v, want code %v", c.target, gotCode, c.wantCode)
		}
		if gotBody != c.wantBody {
			t.Errorf("Requested target %v, got body %v, want body %v", c.target, gotBody, c.wantBody)
		}
	}
}

func TestRedirectHandler(t *testing.T) {
	shorts := map[string]uint64{
		"abcdef": 0,
		"gGG123": 1,
		"qwqwqw": 2,
	}
	file := map[uint64]string{
		0: "/LongURL",
		1: "/ExtraLongURL",
	}

	totalIndex := uint64(3)
	shortURLToIndex := func(str string) (index uint64) {
		index, ok := shorts[str]
		if ok == true {
			return index
		} else {
			return 3
		}
	}
	readFromFile := func(index uint64) (string, error) {
		str, ok := file[index]
		if ok == true {
			return str, nil
		} else {
			return "", errors.New("File read error")
		}
	}

	cases := []struct {
		target            string
		wantCode          int
		wantRedirLocation string
	}{
		// bad input
		{"http://localhost:8080/abcde", 404, "<not tested>"},
		{"http://localhost:8080/abcdefg", 404, "<not tested>"},
		{"http://localhost:8080/ab?def", 404, "<not tested>"},
		// not in database
		{"http://localhost:8080/xyzxyz", 404, "<not tested>"},
		// read error
		{"http://localhost:8080/qwqwqw", http.StatusInternalServerError, "<not tested>"},
		// redirects
		{"http://localhost:8080/abcdef", 302, "/LongURL"},
		{"http://localhost:8080/gGG123", 302, "/ExtraLongURL"},
	}

	for _, c := range cases {
		req := httptest.NewRequest("GET", c.target, nil)
		w := httptest.NewRecorder()
		RedirectHandler(w, req, &totalIndex, shortURLToIndex, readFromFile)

		resp := w.Result()
		gotCode := resp.StatusCode

		if gotCode != c.wantCode {
			t.Errorf("Requested target %v, got code %v, want code %v", c.target, gotCode, c.wantCode)
		}

		if gotCode == 302 {
			loc, _ := resp.Location()
			gotRedirLocation := loc.Path

			if gotRedirLocation != c.wantRedirLocation {
				t.Errorf("Requested target %v, got redirLocation %v, want redirLocation %v",
					c.target, gotRedirLocation, c.wantRedirLocation)
			}
		}
	}
}

func TestShortenHandler(t *testing.T) {
	var msg string
	shorts := map[uint64]string{
		0: "GoLang",
		1: "a1b2c3",
	}

	var totalIndex uint64
	saveToFile := func(index uint64, longURL string) error {
		msg = fmt.Sprintf("Call to saveToFile: index %v, longURL %v", index, longURL)
		if longURL == "fail save" {
			return errors.New("File write error")
		}
		return nil
	}
	indexToShortURL := func(index uint64) string {
		return shorts[index]
	}
	execute := func(wr io.Writer, data interface{}) error {
		value := data.(*struct{ ShortURL, LongURL string })
		fmt.Fprintf(wr, "Short: %v, long: %v", value.ShortURL, value.LongURL)
		if value.LongURL == "fail execute" {
			return errors.New("Template execution error")
		}
		return nil
	}

	cases := []struct {
		target   string
		body     string
		wantCode int
		wantBody string
		wantMsg  string
	}{
		//
		// Testing incomplete. Needs to add more test cases.
		//
		// I could not figure out how to send http requests with request body
		//
		{"http://localhost:8080/shorten/foo", "", 404, "<not tested>", ""},
		{"http://localhost:8080/shorten/", "", http.StatusBadRequest, "<not tested>", ""},
	}

	for _, c := range cases {
		req := httptest.NewRequest("POST", c.target, nil) // TODO add request body
		w := httptest.NewRecorder()
		ShortenHandler(w, req, &totalIndex, saveToFile, indexToShortURL, execute)

		resp := w.Result()
		gotCode := resp.StatusCode
		gotBodyByte, _ := io.ReadAll(resp.Body)
		gotBody := string(gotBodyByte)

		if gotCode != c.wantCode {
			t.Errorf("Requested target %v, got code %v, want code %v", c.target, gotCode, c.wantCode)
		}
		if gotCode == 200 {
			if gotBody != c.wantBody {
				t.Errorf("Requested target %v, got body %v, want body %v", c.target, gotBody, c.wantBody)
			}
		}
		if msg != c.wantMsg {
			t.Errorf("Requested target %v, got msg %v, want msg %v", c.target, msg, c.wantMsg)
		}
	}
}
