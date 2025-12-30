package main

import (
	"fmt"
	"image"
	"os"
	"strconv"
	"strings"

	"github.com/signintech/gopdf"
)

const (
	quantityColumnOffset = 360
	rateColumnOffset     = 405
	amountColumnOffset   = 480
)

const (
	subtotalLabel = "Subtotal"
	discountLabel = "Discount"
	taxLabel      = "Tax"
	totalLabel    = "Total"
)

// Renderer handles PDF rendering
type Renderer struct {
	pdf  *gopdf.GoPdf
	data Invoice
}

// NewRenderer creates a new Renderer
func NewRenderer(pdf *gopdf.GoPdf, data Invoice) *Renderer {
	return &Renderer{pdf: pdf, data: data}
}

// Render generates the full invoice PDF
func (r *Renderer) Render() error {
	r.writeLogo()
	r.writeTitle()
	r.writeBillTo()
	r.writeHeaderRow()

	subtotal := 0.0
	for i := range r.data.Items {
		q := 1
		if len(r.data.Quantities) > i {
			q = r.data.Quantities[i]
		}

		rate := 0.0
		if len(r.data.Rates) > i {
			rate = r.data.Rates[i]
		}

		r.writeRow(r.data.Items[i], q, rate)
		subtotal += float64(q) * rate
	}

	if r.data.Note != "" {
		r.writeNotes()
	}

	r.writeTotals(subtotal, subtotal*r.data.Tax, subtotal*r.data.Discount)

	if r.data.Due != "" {
		r.writeDueDate()
	}

	r.writeFooter()
	return nil
}

func (r *Renderer) writeLogo() {
	if r.data.Logo != "" {
		width, height := getImageDimension(r.data.Logo)
		scaledWidth := 100.0
		scaledHeight := float64(height) * scaledWidth / float64(width)
		_ = r.pdf.Image(r.data.Logo, r.pdf.GetX(), r.pdf.GetY(), &gopdf.Rect{W: scaledWidth, H: scaledHeight})
		r.pdf.Br(scaledHeight + 24)
	}

	r.pdf.SetTextColor(55, 55, 55)
	formatted := strings.ReplaceAll(r.data.From, `\n`, "\n")
	lines := strings.Split(formatted, "\n")
	for i, line := range lines {
		if i == 0 {
			_ = r.pdf.SetFont("Inter", "", 12)
			_ = r.pdf.Cell(nil, line)
			r.pdf.Br(18)
		} else {
			_ = r.pdf.SetFont("Inter", "", 10)
			_ = r.pdf.Cell(nil, line)
			r.pdf.Br(15)
		}
	}
	r.pdf.Br(21)
	r.pdf.SetStrokeColor(225, 225, 225)
	r.pdf.Line(r.pdf.GetX(), r.pdf.GetY(), 260, r.pdf.GetY())
	r.pdf.Br(36)
}

func (r *Renderer) writeTitle() {
	_ = r.pdf.SetFont("Inter-Bold", "", 24)
	r.pdf.SetTextColor(0, 0, 0)
	_ = r.pdf.Cell(nil, r.data.Title)
	r.pdf.Br(36)

	_ = r.pdf.SetFont("Inter", "", 12)
	r.pdf.SetTextColor(100, 100, 100)
	_ = r.pdf.Cell(nil, "#")
	_ = r.pdf.Cell(nil, r.data.Id)
	r.pdf.SetTextColor(150, 150, 150)
	_ = r.pdf.Cell(nil, "  ·  ")
	r.pdf.SetTextColor(100, 100, 100)
	_ = r.pdf.Cell(nil, r.data.Date)
	r.pdf.Br(48)
}

func (r *Renderer) writeDueDate() {
	_ = r.pdf.SetFont("Inter", "", 9)
	r.pdf.SetTextColor(75, 75, 75)
	r.pdf.SetX(rateColumnOffset)
	_ = r.pdf.Cell(nil, "Due Date")
	r.pdf.SetTextColor(0, 0, 0)
	_ = r.pdf.SetFontSize(11)
	r.pdf.SetX(amountColumnOffset - 15)
	_ = r.pdf.Cell(nil, r.data.Due)
	r.pdf.Br(12)
}

func (r *Renderer) writeBillTo() {
	r.pdf.SetTextColor(75, 75, 75)
	_ = r.pdf.SetFont("Inter", "", 9)
	_ = r.pdf.Cell(nil, "BILL TO")
	r.pdf.Br(18)

	formatted := strings.ReplaceAll(r.data.To, `\n`, "\n")
	lines := strings.Split(formatted, "\n")
	for i, line := range lines {
		if i == 0 {
			_ = r.pdf.SetFont("Inter", "", 15)
			_ = r.pdf.Cell(nil, line)
			r.pdf.Br(20)
		} else {
			_ = r.pdf.SetFont("Inter", "", 10)
			_ = r.pdf.Cell(nil, line)
			r.pdf.Br(15)
		}
	}
	r.pdf.Br(64)
}

func (r *Renderer) writeHeaderRow() {
	_ = r.pdf.SetFont("Inter", "", 9)
	r.pdf.SetTextColor(55, 55, 55)
	_ = r.pdf.Cell(nil, "ITEM")
	r.pdf.SetX(quantityColumnOffset)
	_ = r.pdf.Cell(nil, "QTY")
	r.pdf.SetX(rateColumnOffset)
	_ = r.pdf.Cell(nil, "RATE")
	r.pdf.SetX(amountColumnOffset)
	_ = r.pdf.Cell(nil, "AMOUNT")
	r.pdf.Br(24)
}

func (r *Renderer) writeNotes() {
	r.pdf.SetY(600)
	_ = r.pdf.SetFont("Inter", "", 9)
	r.pdf.SetTextColor(55, 55, 55)
	_ = r.pdf.Cell(nil, "NOTES")
	r.pdf.Br(18)
	r.pdf.SetTextColor(0, 0, 0)
	formatted := strings.ReplaceAll(r.data.Note, `\n`, "\n")
	lines := strings.Split(formatted, "\n")
	for _, line := range lines {
		_ = r.pdf.Cell(nil, line)
		r.pdf.Br(15)
	}
	r.pdf.Br(48)
}

func (r *Renderer) writeFooter() {
	r.pdf.SetY(800)
	_ = r.pdf.SetFont("Inter", "", 10)
	r.pdf.SetTextColor(55, 55, 55)
	_ = r.pdf.Cell(nil, r.data.Id)
	r.pdf.SetStrokeColor(225, 225, 225)
	r.pdf.Line(r.pdf.GetX()+10, r.pdf.GetY()+6, 550, r.pdf.GetY()+6)
	r.pdf.Br(48)
}

func (r *Renderer) writeRow(item string, quantity int, rate float64) {
	_ = r.pdf.SetFont("Inter", "", 11)
	r.pdf.SetTextColor(0, 0, 0)

	total := float64(quantity) * rate
	amount := strconv.FormatFloat(total, 'f', 2, 64)

	_ = r.pdf.Cell(nil, item)
	r.pdf.SetX(quantityColumnOffset)
	_ = r.pdf.Cell(nil, strconv.Itoa(quantity))
	r.pdf.SetX(rateColumnOffset)
	_ = r.pdf.Cell(nil, currencySymbols[r.data.Currency]+strconv.FormatFloat(rate, 'f', 2, 64))
	r.pdf.SetX(amountColumnOffset)
	_ = r.pdf.Cell(nil, currencySymbols[r.data.Currency]+amount)
	r.pdf.Br(24)
}

func (r *Renderer) writeTotals(subtotal, tax, discount float64) {
	r.pdf.SetY(600)
	r.writeTotal(subtotalLabel, subtotal)
	if tax > 0 {
		r.writeTotal(taxLabel, tax)
	}
	if discount > 0 {
		r.writeTotal(discountLabel, discount)
	}
	r.writeTotal(totalLabel, subtotal+tax-discount)
}

func (r *Renderer) writeTotal(label string, total float64) {
	_ = r.pdf.SetFont("Inter", "", 9)
	r.pdf.SetTextColor(75, 75, 75)
	r.pdf.SetX(rateColumnOffset)
	_ = r.pdf.Cell(nil, label)
	r.pdf.SetTextColor(0, 0, 0)
	_ = r.pdf.SetFontSize(12)
	r.pdf.SetX(amountColumnOffset - 15)
	if label == totalLabel {
		_ = r.pdf.SetFont("Inter-Bold", "", 11.5)
	}
	_ = r.pdf.Cell(nil, currencySymbols[r.data.Currency]+strconv.FormatFloat(total, 'f', 2, 64))
	r.pdf.Br(24)
}

// getImageDimension returns the width and height of an image file
func getImageDimension(imagePath string) (int, int) {
	file, err := os.Open(imagePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
	}
	defer file.Close()

	img, _, err := image.DecodeConfig(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", imagePath, err)
	}
	return img.Width, img.Height
}
