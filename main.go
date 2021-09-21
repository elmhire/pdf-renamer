package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"unicode"

	"github.com/ledongthuc/pdf"
)

// Test comment
func main() {
	pdf.DebugOn = true
	files := getPdfFiles()
	for _, file := range files {
		invoice, err := readPdf(file) // Read local pdf file
		if err != nil {
			panic(err)
		}
		fmt.Println(
			getBillToName(invoice))
	}
}

func getPdfFiles() []string {
	var pdfList []string
	files, err := ioutil.ReadDir(".")
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range files {
		name := file.Name()
		if strings.Contains(name, ".pdf") {
			pdfList = append(pdfList, name)
		}
	}
	return pdfList
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
