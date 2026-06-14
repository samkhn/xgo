package main

import (
	//"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestListBooks(t *testing.T) {
	store := NewBookStore()
	store.Create(Book{Title: "Test Book", Author: "Test Author"})
	handler := NewBookHandler(store)
	r := chi.NewRouter()
	r.Mount("/books", handler.Routes())
	req := httptest.NewRequest("GET", "/books", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Got %d, want %d", w.Code, http.StatusOK)
	}
	var books []Book
	json.NewDecoder(w.Body).Decode(&books)
	if len(books) != 1 {
		t.Errorf("Got %d book(s) returned, want %d", len(books), 1)
	}
}
