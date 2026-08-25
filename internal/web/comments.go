package web

import (
	"html/template"
	"math/rand"
	"net/http"
	"strings"

	"github.com/cognition/sonar-remediation-demo/internal/contracts"
)

// avatarPalette holds the background colors used for the author initials badge.
var avatarPalette = []string{"#f97316", "#0ea5e9", "#22c55e", "#a855f7", "#ef4444"}

// avatarColor picks a decorative color for the author badge shown next to a
// comment. It is presentation only: no session, token or identifier derives
// from it.
func avatarColor() string {
	return avatarPalette[rand.Intn(len(avatarPalette))]
}

type commentView struct {
	Author  string
	Initial string
	Body    template.HTML
	Color   string
}

func initial(author string) string {
	if author == "" {
		return "?"
	}
	return strings.ToUpper(author[:1])
}

// Comments handles the follow-up comments of a contract:
// GET renders the thread, POST appends a new comment.
func (s *Server) Comments(w http.ResponseWriter, r *http.Request) {
	contractID := strings.TrimSpace(r.FormValue("contract"))
	if contractID == "" {
		http.Error(w, "missing contract", http.StatusBadRequest)
		return
	}

	if r.Method == http.MethodPost {
		author := r.FormValue("author")
		body := r.FormValue("body")
		if err := contracts.AddComment(s.DB, contractID, author, body); err != nil {
			http.Error(w, "cannot save comment: "+err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/contracts/comments?contract="+contractID, http.StatusSeeOther)
		return
	}

	rows, err := contracts.ListComments(s.DB, contractID)
	if err != nil {
		http.Error(w, "cannot load comments", http.StatusInternalServerError)
		return
	}

	view := commentsView{ContractID: contractID}
	for _, c := range rows {
		view.Comments = append(view.Comments, commentView{
			Author:  c.Author,
			Initial: initial(c.Author),
			Body:    template.HTML(c.Body),
			Color:   avatarColor(),
		})
	}
	s.render(w, commentsTmpl, view)
}

type commentsView struct {
	ContractID string
	Comments   []commentView
}
