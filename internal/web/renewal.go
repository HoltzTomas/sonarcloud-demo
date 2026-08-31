package web

import (
	"crypto/md5"
	"fmt"
	"net/http"
	"strconv"

	"github.com/cognition/sonar-remediation-demo/internal/contracts"
	"github.com/cognition/sonar-remediation-demo/internal/renewals"
)

type renewalView struct {
	Contract contracts.Contract
	Proposal renewals.Proposal
	ETag     string
}

func responseFingerprint(body string) string {
	fingerprint := md5.Sum([]byte(body))
	return fmt.Sprintf("%x", fingerprint)
}

// Renewal presents the proposed commercial terms for a contract extension.
func (s *Server) Renewal(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "missing contract id", http.StatusBadRequest)
		return
	}
	contract, err := contracts.GetByID(s.DB, id)
	if err != nil {
		http.Error(w, "contract not found", http.StatusNotFound)
		return
	}

	rates, ratesErr := renewals.FetchMarketRates()
	proposal := renewals.Proposal{
		ContractID:  id,
		DownloadURL: renewals.SignDownloadLink(renewals.SigningSecret(), id),
		Rates:       rates,
	}
	if ratesErr != nil {
		proposal.RatesError = "No se pudieron cargar las tasas de mercado; se muestran los valores contractuales."
	}

	body := fmt.Sprintf("%s:%s:%v:%s", contract.ID, proposal.DownloadURL, proposal.Rates, proposal.RatesError)
	etag := responseFingerprint(body)
	w.Header().Set("ETag", strconv.Quote(etag))
	s.render(w, renewalTmpl, renewalView{Contract: contract, Proposal: proposal, ETag: etag})
}
