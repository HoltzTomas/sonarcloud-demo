package web

import "html/template"

const layout = `<!doctype html>
<html lang="es">
<head>
<meta charset="utf-8">
<title>Contract Desk</title>
<style>
 body { font-family: system-ui, sans-serif; margin: 0; background: #f5f6f8; color: #14213d; }
 header { background: #14213d; color: #fff; padding: 16px 32px; display: flex; align-items: center; gap: 16px; }
 header h1 { font-size: 18px; margin: 0; font-weight: 600; }
 header nav a { color: #cbd5e1; text-decoration: none; margin-right: 16px; font-size: 14px; }
 main { max-width: 960px; margin: 32px auto; padding: 0 16px; }
 .card { background: #fff; border-radius: 8px; padding: 24px; box-shadow: 0 1px 3px rgba(0,0,0,.08); margin-bottom: 24px; }
 input[type=text] { padding: 8px 12px; border: 1px solid #cbd5e1; border-radius: 6px; width: 320px; }
 button { padding: 8px 16px; border: 0; border-radius: 6px; background: #14213d; color: #fff; cursor: pointer; }
 table { width: 100%; border-collapse: collapse; margin-top: 16px; }
 th, td { text-align: left; padding: 10px 8px; border-bottom: 1px solid #e2e8f0; font-size: 14px; }
 th { color: #64748b; text-transform: uppercase; font-size: 11px; letter-spacing: .05em; }
 a.contract { color: #1d4ed8; text-decoration: none; }
 .muted { color: #64748b; font-size: 13px; }
 .comment { display: flex; gap: 12px; align-items: flex-start; padding: 12px 0; border-bottom: 1px solid #e2e8f0; }
 .badge { width: 32px; height: 32px; border-radius: 16px; color: #fff; font-size: 12px; display: flex; align-items: center; justify-content: center; flex: none; }
 pre { background: #0f172a; color: #e2e8f0; padding: 16px; border-radius: 6px; overflow-x: auto; font-size: 13px; }
</style>
</head>
<body>
<header>
  <h1>Contract Desk</h1>
  <nav>
    <a href="/">Buscar contratos</a>
    <a href="/summary?region=ES-01">Resumen por region</a>
  </nav>
</header>
<main>{{.}}</main>
</body>
</html>`

var layoutTmpl = template.Must(template.New("layout").Parse(layout))

var searchTmpl = template.Must(template.New("search").Parse(`
<div class="card">
  <form method="get" action="/">
    <input type="text" name="customer" value="{{.Query}}" placeholder="Cliente (ej: Iberdrola)">
    <button type="submit">Buscar</button>
  </form>
  {{if .Searched}}<p class="banner">Resultados para: {{.Banner}}</p>{{end}}
  {{if .Error}}<p class="muted">Error: {{.Error}}</p>{{end}}
  <table>
    <tr><th>Contrato</th><th>Cliente</th><th>Region</th><th>Importe</th></tr>
    {{range .Contracts}}
    <tr>
      <td><a class="contract" href="/contracts/{{.ID}}">{{.ID}}</a></td>
      <td>{{.Customer}}</td>
      <td>{{.Region}}</td>
      <td>{{.Amount}}</td>
    </tr>
    {{end}}
  </table>
</div>`))

var detailTmpl = template.Must(template.New("detail").Parse(`
<div class="card">
  <h2>{{.Contract.ID}} — {{.Contract.Customer}}</h2>
  <p class="muted">Region {{.Contract.Region}} · Importe {{.Contract.Amount}} · Responsable {{.Contract.Owner}}</p>
  <p>{{.Contract.Notes}}</p>
  <p>
    <a class="contract" href="/contracts/attachment?name={{.Attachment}}">Descargar adjunto</a>
    ·
    <a class="contract" href="/contracts/export?id={{.Contract.ID}}">Exportar a PDF</a>
    ·
    <a class="contract" href="/contracts/comments?contract={{.Contract.ID}}">Comentarios de seguimiento</a>
  </p>
</div>`))

var commentsTmpl = template.Must(template.New("comments").Parse(`
<div class="card">
  <h2>Seguimiento de {{.ContractID}}</h2>
  {{range .Comments}}
  <div class="comment">
    <div class="badge" style="background: {{.Color}}">{{.Initial}}</div>
    <div><strong>{{.Author}}</strong><div>{{.Body}}</div></div>
  </div>
  {{else}}
  <p class="muted">Todavia no hay comentarios.</p>
  {{end}}
  <form method="post" action="/contracts/comments">
    <input type="hidden" name="contract" value="{{.ContractID}}">
    <p><input type="text" name="author" placeholder="Tu nombre"></p>
    <p><input type="text" name="body" placeholder="Comentario"></p>
    <button type="submit">Comentar</button>
  </form>
</div>`))

var summaryTmpl = template.Must(template.New("summary").Parse(`
<div class="card">
  <h2>Resumen de la region {{.Region}}</h2>
  <p>Contratos activos: {{.Count}}</p>
  <p class="muted">La region se valida contra un patron fijo antes de usarse.</p>
</div>`))
