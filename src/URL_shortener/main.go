package main

import (
	"URL_shortener/SURLTools"
	"encoding/csv"
	"errors"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
)

const fileName string = "savedURLs.csv"

var tmpl = template.Must(template.ParseFiles("success.html"))
var validShortURL = regexp.MustCompile("^[0-9a-zA-Z]{6}$")
var TotalIndex uint64

func ReadFromFile(index uint64) (string, error) {
	f, err := os.Open(fileName)
	if err != nil {
		log.Println(err)
		return "", err
	}
	defer f.Close()
	r := csv.NewReader(f)
	for i := uint64(0); i < index; i++ {
		r.Read()
		// read error is caught below by comparing retrievedIndex to index
	}
	record, err := r.Read()
	if err != nil {
		log.Println(err)
		return "", err
	}
	retrievedIndex, err := strconv.ParseUint(record[0], 10, 64)
	if err != nil {
		log.Fatal("Error: csv file corrupted\n", err)
	}
	if retrievedIndex != index {
		err = errors.New("read error")
		log.Println(err)
		return "", err
	}
	return record[1], nil
}

func SaveToFile(index uint64, longURL string) error {
	f, err := os.OpenFile(fileName, os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		log.Println(err)
		return err
	}
	defer f.Close()
	w := csv.NewWriter(f)
	record := []string{strconv.FormatUint(index, 10), longURL}
	err = w.Write(record)
	if err != nil {
		log.Println(err)
		return err
	}
	w.Flush()
	return nil
}

func MakeHandler(fn func(http.ResponseWriter, *http.Request, ...interface{})) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fn(w, r)
	}
}

func DetermineHandler(w http.ResponseWriter, r *http.Request, mock ...interface{}) {
	indexPageHandler := MakeHandler(IndexPageHandler)
	redirectHandler := MakeHandler(RedirectHandler)
	if mock != nil {
		indexPageHandler = mock[0].(func(w http.ResponseWriter, r *http.Request))
		redirectHandler = mock[1].(func(w http.ResponseWriter, r *http.Request))
	}

	if r.URL.Path == "" || r.URL.Path == "/" || r.URL.Path == "/index.html" {
		indexPageHandler(w, r)
	} else {
		redirectHandler(w, r)
	}
}

func IndexPageHandler(w http.ResponseWriter, r *http.Request, mock ...interface{}) {
	serveFileFunc := http.ServeFile
	if mock != nil {
		serveFileFunc = mock[0].(func(w http.ResponseWriter, r *http.Request, name string))
	}

	serveFileFunc(w, r, "index.html")
}

func RedirectHandler(w http.ResponseWriter, r *http.Request, mock ...interface{}) {
	totalIndex := &TotalIndex
	shortURLToIndex := SURLTools.ShortURLToIndex
	readFromFile := ReadFromFile
	if mock != nil {
		totalIndex = mock[0].(*uint64)
		shortURLToIndex = mock[1].(func(str string) (index uint64))
		readFromFile = mock[2].(func(index uint64) (string, error))
	}

	shortURL := r.URL.Path[1:]
	if validShortURL.MatchString(shortURL) == false {
		http.NotFound(w, r)
		return
	}
	index := shortURLToIndex(shortURL)
	if index >= *totalIndex {
		http.NotFound(w, r)
		return
	}
	longURL, err := readFromFile(index)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, longURL, http.StatusFound)
}

func ShortenHandler(w http.ResponseWriter, r *http.Request, mock ...interface{}) {
	totalIndex := &TotalIndex
	saveToFile := SaveToFile
	indexToShortURL := SURLTools.IndexToShortURL
	execute := tmpl.Execute
	if mock != nil {
		totalIndex = mock[0].(*uint64)
		saveToFile = mock[1].(func(index uint64, longURL string) error)
		indexToShortURL = mock[2].(func(index uint64) string)
		execute = mock[3].(func(wr io.Writer, data interface{}) error)
	}

	if r.URL.Path != "/shorten/" {
		http.NotFound(w, r)
		return
	}
	longURL := r.FormValue("body")
	if longURL == "" {
		http.Error(w, "no URL to shorten", http.StatusBadRequest)
		return
	}
	if *totalIndex >= SURLTools.MAX_URLS {
		http.Error(
			w,
			"set of short URLs exhausted",
			http.StatusInternalServerError,
		)
		return
	}
	// TODO if short URL already exists, do not create new entry
	err := saveToFile(*totalIndex, longURL)
	if err != nil {
		http.Error(w, "unable to save URL", http.StatusInternalServerError)
	}
	shortURL := indexToShortURL(*totalIndex)
	(*totalIndex)++
	data := &struct{ ShortURL, LongURL string }{shortURL, longURL}
	err = execute(w, data)
	if err != nil {
		http.Error(w, "unable to render page", http.StatusInternalServerError)
	}
}

func main() {
	// the purpose of the whole following block is to load TotalIndex
	f, err := os.Open(fileName)
	if err != nil {
		log.Println("File " + fileName + " not opened. Check if it is not" +
			" missing. If it does not exist, create it.")
		log.Fatal(err)
		// we can not automatically create a new file, because it may already
		// exist and we could overwrite it, destroying all saved data
	}
	r := csv.NewReader(f)
	record := make([]string, 2)
	for true {
		nr, err := r.Read()
		if err == io.EOF {
			break
		}
		record = nr
		if err != nil {
			log.Println("Error while initializing server. Please restart the" +
				" server.")
			log.Fatal(err)
		}
	}
	f.Close()
	if record[0] == "" {
		TotalIndex = 0
	} else {
		lastIndex, err := strconv.ParseUint(record[0], 10, 64)
		if err != nil {
			log.Fatal("Error: csv file corrupted\n", err)
		}
		TotalIndex = lastIndex + 1
	}
	// TotalIndex is finally loaded

	// the first http handler may be used either for redirection or for visiting
	// the index page
	http.HandleFunc("/", MakeHandler(DetermineHandler))

	http.HandleFunc("/shorten/", MakeHandler(ShortenHandler))

	err = http.ListenAndServe(":8080", nil)
	log.Fatal(err)
}
