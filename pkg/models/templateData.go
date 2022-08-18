package models

// TemplateData holds data send from handlers to template
type TemplateData struct {
	StringMap map[string]string
	IntMap    map[string]int
	FloatMap  map[string]float64
	Data      map[string]interface{}
	CSRF      string
	Flash     string
	Warning   string
	Error     string
}
