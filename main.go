package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

func main() {
	html := getInvoiceHTML()
	mockServe := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(html)
	}))
	ctx := context.Background()
	var pdfContent []byte
	err := printPdf(ctx, &pdfContent, mockServe.URL)
	if err != nil {
		log.Fatal("printPdf", err)
	}
	savePDF(pdfContent)
	// http.ListenAndServe(":3000", newAPI(pdfContent))
}

func newAPI(pdfContent []byte) *http.ServeMux {
	serve := http.NewServeMux()
	serve.Handle("/invoice", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/pdf")
		w.Write(pdfContent)
	}))
	return serve
}

func getInvoiceHTML() []byte {
	file, err := os.Open("invoice.html")
	if err != nil {
		log.Fatalln(err)
	}
	defer file.Close()
	stat, err := file.Stat()
	if err != nil {
		fmt.Println(err)
	}
	// Read the file into a byte slice
	f := make([]byte, stat.Size())
	_, err = bufio.NewReader(file).Read(f)
	if err != nil && err != io.EOF {
		fmt.Println(err)
	}
	return f
}

func printPdf(ctx context.Context, res *[]byte, url string) error {
	options := []chromedp.ExecAllocatorOption{
		chromedp.DisableGPU,
		chromedp.NoSandbox,
		chromedp.Headless,
		chromedp.Flag("no-zygote", true),
	}
	cctx, cancel := chromedp.NewExecAllocator(ctx, options...)
	defer cancel()
	cctx, cancel = chromedp.NewContext(cctx)
	defer cancel()
	footerHTML := `
	<span style='width: 100%; text-align: right; font-size: 10px;padding-right:24px;'>
		<span class='pageNumber'></span>/<span class='totalPages'></span>
	</span>`
	return chromedp.Run(cctx, chromedp.Tasks{
		chromedp.Navigate(url),
		chromedp.ActionFunc(func(ctx context.Context) error {
			buf, _, err := page.PrintToPDF().
				WithPaperHeight(11.7).
				WithPaperWidth(8.3).
				WithScale(0.62).
				WithDisplayHeaderFooter(true).
				WithHeaderTemplate(`<style>body { margin: 0; }</style>`).
				WithFooterTemplate(footerHTML).
				Do(ctx)
			if err != nil {
				return err
			}
			*res = buf
			return nil
		}),
	})
}

func savePDF(pdfContent []byte) {
	fe, err := os.Create("invoice.pdf")
	if err != nil {
		log.Println("os.Create", err)
	}
	defer fe.Close()
	_, err = fe.Write(pdfContent)
	if err != nil {
		log.Println("os.Write", err)
	}
}
