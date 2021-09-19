package main

import (
	"bytes"
	"fmt"
	"strings"
	"unicode"

	"github.com/ledongthuc/pdf"
)

func main() {
	pdf.DebugOn = true
	invoice, err := readPdf("test.pdf") // Read local pdf file
	if err != nil {
		panic(err)
	}
	fmt.Println(invoice)

	billToName := getBillToName(invoice)

	fmt.Println(billToName)
}

func readPdf(path string) (string, error) {
	f, r, err := pdf.Open(path)
	// remember close file
	defer f.Close()
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	b, err := r.GetPlainText()
	if err != nil {
		return "", err
	}
	buf.ReadFrom(b)
	return buf.String(), nil
}

func getBillToName(pdfStr string) string {
	pdfStr = pdfStr[strings.Index(pdfStr, "BILL TO:")+8:]

	for i, character := range pdfStr {
		if unicode.IsDigit(character) {
			return strings.Replace(pdfStr[:i], " ", "_", -1)
		}
	}
	return strings.Replace(pdfStr, " ", "_", -1)
}
