package main

import (
	"bytes"
	"html/template"
	"log/slog"
	"net/http"
)

var (
	// Pre-parsed templates avoid disk I/O and parsing on every request.
	indexTemplate = template.Must(template.ParseFiles("./views/index.html"))
	leftTemplate  = template.Must(template.ParseFiles("./views/left.html"))
)

var monthNames = [...]string{
	"",
	"Enero",
	"Febrero",
	"Marzo",
	"Abril",
	"Mayo",
	"Junio",
	"Julio",
	"Agosto",
	"Septiembre",
	"Octubre",
	"Noviembre",
	"Diciembre",
}

var weekdayNames = [...]string{
	"Domingo",
	"Lunes",
	"Martes",
	"Miércoles",
	"Jueves",
	"Viernes",
	"Sábado",
}

func renderTemplate(w http.ResponseWriter, tmpl *template.Template, data any) {
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		slog.Error("failed to execute template", "template", tmpl.Name(), "error", err)
		writeTextError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeHTML(w, http.StatusOK, buf.Bytes())
}
