package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Book struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Author      string    `json:"author"`
	ISBN        string    `json:"isbn"`
	PublishedAt time.Time `json:"published_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type BookStore struct {
	// TODO: maybe replace with channel?
	mu     sync.RWMutex
	books  map[string]Book
	nextID int
}

func NewBookStore() *BookStore {
	return &BookStore{
		books:  make(map[string]Book),
		nextID: 1,
	}
}

func (s *BookStore) GetAll() []Book {
	s.mu.RLock()
	defer s.mu.RUnlock()
	books := make([]Book, 0, len(s.books))
	for _, book := range s.books {
		books = append(books, book)
	}
	return books
}

func (s *BookStore) Get(id string) (Book, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	book, ok := s.books[id]
	return book, ok
}

func (s *BookStore) Create(book Book) Book {
	s.mu.Lock()
	defer s.mu.Unlock()

	book.ID = fmt.Sprintf("%d", s.nextID)
	s.nextID++
	book.CreatedAt = time.Now()
	book.UpdatedAt = time.Now()
	s.books[book.ID] = book
	return book
}

func (s *BookStore) Update(id string, book Book) (Book, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if existing, ok := s.books[id]; ok {
		book.ID = id
		book.CreatedAt = existing.CreatedAt
		book.UpdatedAt = time.Now()
		s.books[id] = book
		return book, true
	}
	return Book{}, false
}

func (s *BookStore) Delete(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.books[id]; ok {
		delete(s.books, id)
		return true
	}
	return false
}

type BookHandler struct {
	store *BookStore
}

func NewBookHandler(store *BookStore) *BookHandler {
	return &BookHandler{store: store}
}

func (h *BookHandler) BookCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bookID := chi.URLParam(r, "bookID")
		if bookID != "" {
			book, ok := h.store.Get(bookID)
			if !ok {
				http.Error(w, "book not found", http.StatusNotFound)
				return
			}
			ctx := context.WithValue(r.Context(), "book", book)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		next.ServeHTTP(w, r)
	})
}

func respondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}

func (h *BookHandler) List(w http.ResponseWriter, r *http.Request) {
	books := h.store.GetAll()
	respondJSON(w, http.StatusOK, books)
}

func (h *BookHandler) Get(w http.ResponseWriter, r *http.Request) {
	book := r.Context().Value("book").(Book)
	respondJSON(w, http.StatusOK, book)
}

func (h *BookHandler) Create(w http.ResponseWriter, r *http.Request) {
	var book Book
	if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if book.Title == "" || book.Author == "" {
		respondError(w, http.StatusBadRequest, "title and author required")
		return
	}
	created := h.store.Create(book)
	respondJSON(w, http.StatusCreated, created)
}

func (h *BookHandler) Update(w http.ResponseWriter, r *http.Request) {
	bookID := chi.URLParam(r, "bookID")
	var book Book
	if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	updated, ok := h.store.Update(bookID, book)
	if !ok {
		respondError(w, http.StatusBadRequest, "book not found")
		return
	}
	respondJSON(w, http.StatusOK, updated)
}

func (h *BookHandler) Delete(w http.ResponseWriter, r *http.Request) {
	bookID := chi.URLParam(r, "bookID")
	if !h.store.Delete(bookID) {
		respondError(w, http.StatusBadRequest, "book not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *BookHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Use(h.BookCtx)
	r.Get("/", h.List)
	r.Post("/", h.Create)
	r.Route("/{bookID}", func(r chi.Router) {
		r.Get("/", h.Get)
		r.Put("/", h.Update)
		r.Delete("/", h.Delete)
	})

	return r
}

func main() {
	store := NewBookStore()
	store.Create(Book{
		Title:       "The Go Programming Language",
		Author:      "Alan Donovan & Brian Kernighan",
		ISBN:        "978-0134190440",
		PublishedAt: time.Date(2015, 10, 26, 0, 0, 0, 0, time.UTC),
	})
	store.Create(Book{
		Title:       "Learning Go",
		Author:      "Jon Bodner",
		ISBN:        "978-1492077213",
		PublishedAt: time.Date(2021, 3, 23, 0, 0, 0, 0, time.UTC),
	})

	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Get("/statusz", func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusOK, map[string]string{
			"status": "healthy",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	r.Route("/api/v1", func(r chi.Router) {
		bookHandler := NewBookHandler(store)
		r.Mount("/books", bookHandler.Routes())
	})

	addr := ":3000"
	log.Printf("Starting server on %s", addr)
	log.Fatal(http.ListenAndServe(":3000", r))
}
