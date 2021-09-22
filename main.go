package main

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/DusanKasan/parsemail"
	"github.com/ledongthuc/pdf"
	"github.com/timakin/gonvert"
	"golang.org/x/net/html"
)

func check(err error) {
	if err != nil {
		fmt.Printf("Failed in check(): %s\n", err)
		pause()
		os.Exit(1)
	}
}

func main() {
	// Get email file name input from user
	filename := getFileName()

	// Get data from file
	fmt.Printf("Opening %s...\n\n", filename)
	data := getRawFileDataAsStr(filename)
	content := convertToUTF8(data)

	// Extract links from contents
	links, err := extractLinks(content)
	if err != nil {
		fmt.Println("No files found to download.")
		normalExit()
	}

	// Download links
	fmt.Printf("%d file's to download...\n", len(links))
	complete := downloadLinks(links)
	fmt.Printf("%d of %d files downloaded completely.\n\n", complete, len(links))

	fmt.Println("Renaming files...")
	renameFiles()
	normalExit()
} // End main

func getCwdFileList() []string {
	var (
		listing []string
		fTypes  = []string{"htm", "html", "eml"}
	)

	list, err := os.ReadDir(".")
	check(err)

	for _, f := range list {
		if filepath.Base(os.Args[0]) != f.Name() && !f.IsDir() && existsIn(fTypes, f.Name()) {
			listing = append(listing, f.Name())
		}
	}
	return listing
}

func existsIn(list []string, a string) bool {
	for _, b := range list {
		if strings.Contains(a, strings.ToLower(b)) {
			return true
		}
	}
	return false
}

func getFileName() string {
	files := getCwdFileList()
	var choice string
	for {
		fmt.Println("Please select the email to open.")
		fmt.Println()
		for i, file := range files {
			fmt.Printf("(%d) %s\n", i+1, file)
		}

		reader := bufio.NewReader(os.Stdin)
		fmt.Printf("\n (q to exit)-> ")
		input, _ := reader.ReadString('\n')

		input = strings.TrimSpace(input)
		if strings.ToLower(input) == "q" {
			fmt.Println("Exiting...")
			pause()
			os.Exit(0)
		}
		ans, err := strconv.ParseUint(input, 10, 16)
		if err != nil || int(ans) >= len(files)+1 || int(ans) < 1 {
			fmt.Println("\nInvalid entry, please try again.")
			fmt.Println()
			continue
		} else {
			fmt.Printf("\nYou selected: %v \n\n", ans)
			choice = files[ans-1]
			break
		}
	}
	return choice
}

func getRawFileDataAsStr(filename string) string {
	var data string

	if strings.Contains(filename, "eml") {
		data = getEMLContent(filename)
	} else {
		data = getHTMLContent(filename)
	}
	return string(data)
}

func getEMLContent(filename string) string {
	rd := getBytesReaderFromFile(filename)

	email, err := parsemail.Parse(rd)
	if err == io.EOF {
		fmt.Println("Empty file:", filename, "\nExiting.")
		pause()
		os.Exit(0)
	} else if err != nil {
		fmt.Println("Error reading file:", filename)
		fmt.Println("Either it is corrupt or it is not an 'eml' file.")
		pause()
		os.Exit(1)
	}
	html := email.HTMLBody
	var retVal string
	decoded, err := base64.StdEncoding.DecodeString(html)
	if err != nil {
		fmt.Println("Data not base64 encoded.")
		fmt.Println()
		retVal = html
	} else {
		fmt.Println("Data is base64 encoded.")
		fmt.Println()
		retVal = string(decoded)
	}
	return retVal
}

func getHTMLContent(filename string) string {
	temp, err := os.ReadFile(filename)
	check(err)
	return string(temp)
}

func convertToUTF8(data string) string {
	content, err := gonvert.New(data, gonvert.UTF8).Convert()
	if err != nil {
		fmt.Println("Failed to Convert: ", err)
		pause()
		os.Exit(1)
	}
	return content
}

func getBytesReaderFromFile(filename string) *bytes.Reader {
	dat, err := os.ReadFile(filename)
	if err != nil {
		log.Printf("Failed opening %s: %s", filename, err)
		pause()
		os.Exit(1)
	}
	return bytes.NewReader(dat)
}

// Link ...
type Link struct {
	url  string
	text string
}

func extractLinks(htmlText string) ([]Link, error) {
	var links []Link

	node, err := html.Parse(strings.NewReader(htmlText))
	check(err)

	// Algorithm taken from https://pkg.go.dev/golang.org/x/net/html#pkg-functions
	// Takes an html.Node and walks the tree recursively
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, a := range n.Attr {
				if a.Key == "href" &&
					!strings.Contains(a.Val, "mailto") &&
					!strings.Contains(n.FirstChild.Data, "here") {
					link := Link{
						url:  a.Val,
						text: n.FirstChild.Data,
					}
					links = append(links, link)
					break
				} // End if
			} // End For
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(node)
	if links == nil {
		return nil, errors.New("no links found")
	}
	return links, nil
}

func downloadLinks(links []Link) int {
	complete := 0
	for _, link := range links {
		fmt.Printf("Downloading %s...", link.text)
		err := downloadFile(link, true)
		if err != nil {
			fmt.Printf("\terror, incomplete!\n")
			continue
		}
		fmt.Printf("\tcomplete.\n")
		complete++
	}
	return complete
}

func downloadFile(link Link, useLinkName bool, pathOptional ...string) error {
	var (
		path     string
		filename string
	)

	if len(pathOptional) > 0 {
		path = pathOptional[0]
	} else {
		path = "."
	}

	response, err := http.Get(link.url)
	check(err)

	defer response.Body.Close()

	// Decide whether to use link.text as filename
	// or the name that the server gives us
	if useLinkName {
		filename = fmt.Sprintf("%s/%s", path, link.text)
	} else {
		filename = filepath.Base(response.Request.URL.EscapedPath())
		temp, err := url.PathUnescape(filename)
		if err == nil {
			filename = temp
		}
	}

	// Create the file
	out, err := os.Create(fmt.Sprintf("%s/%s", path, filename))
	if err != nil {
		return err
	}

	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, response.Body)
	out.Close()

	return err
}

func renameFiles() {
	pdf.DebugOn = true

	files := getPdfFiles()
	for _, file := range files {
		invoice, err := readPdf(file) // Read local pdf file
		if err != nil {
			panic(err)
		}
		billToName := getBillToName(invoice)
		src, dest := file, strings.Join([]string{billToName, file}, "-")
		fmt.Println("Renaming: ", src, "to", dest)
		os.Rename(src, dest)
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
		if strings.HasPrefix(name, "SI-") && strings.Contains(name, ".pdf") {
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
	var re = regexp.MustCompile(`('|,)`)
	pdfStr = pdfStr[strings.Index(pdfStr, "BILL TO:")+8:]
	for i, character := range pdfStr {
		if unicode.IsDigit(character) {
			pdfStr = strings.Replace(pdfStr[:i], " ", "_", -1)
			break
		}
	}
	pdfStr = re.ReplaceAllString(pdfStr, ``)
	return pdfStr
}

func normalExit() {
	fmt.Print("\nPress 'Enter' to continue...")
	pause()
	os.Exit(0)
}

func pause() {
	_, err := bufio.NewReader(os.Stdin).ReadBytes('\n')
	if err != nil {
		log.Fatal("Something went wrong: ", err)
	}
}
