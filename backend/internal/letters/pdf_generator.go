package letters

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/jung-kurt/gofpdf"
)

type PDFGenerator struct{}

func NewPDFGenerator() *PDFGenerator {
	return &PDFGenerator{}
}

type LetterData struct {
	CitizenName  string
	ProgramName  string
	BenefitAmount string
	CaseNumber   string
	DenialReason string
	AgencyName   string
}

func (g *PDFGenerator) RenderTemplate(bodyTemplate string, data LetterData) (string, error) {
	tmpl, err := template.New("letter").Parse(bodyTemplate)
	if err != nil {
		return "", fmt.Errorf("parse template: %w", err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}
	return buf.String(), nil
}

func (g *PDFGenerator) GeneratePDF(title, body string) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "Letter", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(40, 10, title)
	pdf.Ln(12)
	pdf.SetFont("Arial", "", 12)

	for _, line := range strings.Split(body, "\n") {
		pdf.MultiCell(0, 6, line, "", "", false)
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("output pdf: %w", err)
	}
	return buf.Bytes(), nil
}

func (g *PDFGenerator) GenerateLetter(title, bodyTemplate string, data LetterData) ([]byte, error) {
	body, err := g.RenderTemplate(bodyTemplate, data)
	if err != nil {
		return nil, err
	}
	return g.GeneratePDF(title, body)
}
