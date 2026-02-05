package main

import (
	"html/template"
	"io"
	"net/http"
	"strings"
)

const maxUploadSize = 10 << 20 // 10MB

type Handler struct {
	storage  *Storage
	template *template.Template
}

func NewHandler(storage *Storage) *Handler {
	tmpl := template.Must(template.ParseFiles("templates/index.html"))
	return &Handler{
		storage:  storage,
		template: tmpl,
	}
}

func (h *Handler) Index(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if path == "/" {
		h.template.Execute(w, nil)
		return
	}

	slug := strings.TrimPrefix(path, "/")
	if slug == "" {
		http.NotFound(w, r)
		return
	}

	content, err := h.storage.Get(slug)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(content)
}

type UploadResult struct {
	Success bool
	URL     string
	Error   string
}

func (h *Handler) Upload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		h.renderResult(w, UploadResult{Error: "File too large (max 10MB)"})
		return
	}

	username := r.FormValue("username")
	if err := h.storage.ValidateUsername(username); err != nil {
		h.renderResult(w, UploadResult{Error: err.Error()})
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		h.renderResult(w, UploadResult{Error: "No file uploaded"})
		return
	}
	defer file.Close()

	if !strings.HasSuffix(strings.ToLower(header.Filename), ".html") {
		h.renderResult(w, UploadResult{Error: "Only HTML files allowed"})
		return
	}

	// Read file content with size limit
	content := io.LimitReader(file, maxUploadSize)

	slug, err := h.storage.Save(username, content)
	if err != nil {
		h.renderResult(w, UploadResult{Error: "Failed to save file"})
		return
	}

	url := "https://insight.funq.kr/" + slug
	h.renderResult(w, UploadResult{Success: true, URL: url})
}

func (h *Handler) renderResult(w http.ResponseWriter, result UploadResult) {
	h.template.Execute(w, result)
}
