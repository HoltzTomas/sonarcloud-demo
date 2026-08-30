package web

import (
	"bytes"
	"database/sql"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/cognition/sonar-remediation-demo/internal/contracts"
)

// AttachmentRoot is the directory that holds contract attachments.
const AttachmentRoot = "data/attachments"

var regionPattern = regexp.MustCompile(`^[A-Z]{2}-[0-9]{2}$`)

// Server serves the contract desk UI.
type Server struct {
	DB *sql.DB
}

// Routes builds the HTTP router of the contract desk.
func (s *Server) Routes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.Search)
	mux.HandleFunc("/contracts/", s.Detail)
	mux.HandleFunc("/contracts/attachment", s.Attachment)
	mux.HandleFunc("/contracts/export", s.Export)
	mux.HandleFunc("/contracts/comments", s.Comments)
	mux.HandleFunc("/summary", s.Summary)
	return mux
}

func (s *Server) render(w http.ResponseWriter, tmpl *template.Template, data interface{}) {
	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		http.Error(w, "render error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := layoutTmpl.Execute(w, template.HTML(body.String())); err != nil {
		http.Error(w, "render error", http.StatusInternalServerError)
	}
}

type searchView struct {
	Query     string
	Banner    template.HTML
	Searched  bool
	Error     string
	Contracts []contracts.Contract
}

// Search renders the contract search page.
func (s *Server) Search(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	customer := r.URL.Query().Get("customer")
	view := searchView{
		Query:    customer,
		Banner:   template.HTML(customer),
		Searched: customer != "",
	}

	if customer != "" {
		rows, err := contracts.FindByCustomer(s.DB, customer)
		if err != nil {
			view.Error = err.Error()
		}
		view.Contracts = rows
	}
	s.render(w, searchTmpl, view)
}

type detailView struct {
	Contract   contracts.Contract
	Attachment string
}

// Detail renders a single contract.
func (s *Server) Detail(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/contracts/")
	if id == "" {
		http.NotFound(w, r)
		return
	}
	c, err := contracts.GetByID(s.DB, id)
	if err != nil {
		http.Error(w, "contract not found", http.StatusNotFound)
		return
	}
	s.render(w, detailTmpl, detailView{Contract: c, Attachment: c.ID + ".txt"})
}

// Attachment streams a contract attachment from disk.
func (s *Server) Attachment(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	data, err := ioutil.ReadFile(filepath.Join(AttachmentRoot, name))
	if err != nil {
		http.Error(w, "attachment not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write(data)
}

// Export renders a contract through the reporting CLI.
func (s *Server) Export(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	out, err := exec.Command("sh", "-c", "scripts/report-cli.sh --contract "+id+" --format pdf").CombinedOutput()
	if err != nil {
		http.Error(w, "export failed: "+string(out), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write(out)
}

type summaryView struct {
	Region string
	Count  int
}

// Summary shows how many contracts a region holds.
func (s *Server) Summary(w http.ResponseWriter, r *http.Request) {
	region := r.URL.Query().Get("region")
	if !regionPattern.MatchString(region) {
		http.Error(w, fmt.Sprintf("invalid region code %q, expected format ES-01", region), http.StatusBadRequest)
		return
	}
	n, err := contracts.CountByValidatedRegion(s.DB, region)
	if err != nil {
		http.Error(w, "summary failed", http.StatusInternalServerError)
		return
	}
	s.render(w, summaryTmpl, summaryView{Region: region, Count: n})
}
