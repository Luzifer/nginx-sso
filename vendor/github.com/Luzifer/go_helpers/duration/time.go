package duration

import (
	"bytes"
	"math"
	"strings"
	"text/template"
	"time"

	"github.com/leekchan/gtf"
)

const defaultDurationFormat = `{{if gt .Years 0}}{{.Years}} year{{.Years|pluralize "s"}}, {{end}}` +
	`{{if gt .Days 0}}{{.Days}} day{{.Days|pluralize "s"}}, {{end}}` +
	`{{if gt .Hours 0}}{{.Hours}} hour{{.Hours|pluralize "s"}}, {{end}}` +
	`{{if gt .Minutes 0}}{{.Minutes}} minute{{.Minutes|pluralize "s"}}, {{end}}` +
	`{{if gt .Seconds 0}}{{.Seconds}} second{{.Seconds|pluralize "s"}}{{end}}`

func HumanizeDuration(in time.Duration) string {
	f, err := CustomHumanizeDuration(in, defaultDurationFormat)
	if err != nil {
		panic(err)
	}
	return strings.Trim(f, " ,")
}

func CustomHumanizeDuration(in time.Duration, tpl string) (string, error) {
	result := struct{ Years, Days, Hours, Minutes, Seconds int64 }{}

	in = time.Duration(math.Abs(float64(in)))

	for in > 0 {
		switch {
		case in > 365.25*24*time.Hour:
			result.Years = int64(in / (365 * 24 * time.Hour))
			in = in - time.Duration(result.Years)*365*24*time.Hour
		case in > 24*time.Hour:
			result.Days = int64(in / (24 * time.Hour))
			in = in - time.Duration(result.Days)*24*time.Hour
		case in > time.Hour:
			result.Hours = int64(in / time.Hour)
			in = in - time.Duration(result.Hours)*time.Hour
		case in > time.Minute:
			result.Minutes = int64(in / time.Minute)
			in = in - time.Duration(result.Minutes)*time.Minute
		default:
			result.Seconds = int64(in / time.Second)
			in = 0
		}
	}

	tmpl, err := template.New("timeformat").Funcs(template.FuncMap(gtf.GtfFuncMap)).Parse(tpl)
	if err != nil {
		return "", err
	}

	buf := bytes.NewBuffer([]byte{})
	tmpl.Execute(buf, result)

	return buf.String(), nil
}
