[Golang]สร้าง PDF ด้วย chromedp
คุณกำลังหาวิธีสร้าง PDF ของภาษา GO อยู่หรือเปล่า คงใช่แหละไม่งั้นคงไม่เขามาหรอกใช่ไหม 55555
เหมือนกันครับช่วงนี้ระบบต้องสร้างไฟล์ PDF ก็เลยมาเขียนบันทึกไว้สักหน่อย จะไม่อธิบายเยอะน่ะครับหาก
อยากอ่าน code เลยเชิญที่นี้ [คลิกนี้](https://github.com/smalldog124/go-gen-pdf)

โดนในที่นี้ใช้ lib [chromedp](https://github.com/chromedp/chromedp)
หรือมีอีกตัวที่เคยใช้ https://medium.com/@smalldoc124/golang-สร้าง-pdf-ภาษาไทย-a2ee0ca86668
1.สร้างไฟล์ html ที่จะ gen PDF
2.อ่านไฟล์ template html
3.สร้างไฟล์ PDF
4.บันทึกไฟล์ / response APIs

1. สร้างไฟล์ html ที่จะ gen PDF
https://gist.github.com/smalldog124/7f3c16a5fb663aea6bfdbff516e8a446
2. อ่านไฟล์ template html
```go
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
```
3.สร้างไฟล์ PDF
```go
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
```
4.บันทึกไฟล์ / response APIs
บันทืกไฟล์
```go
    fe, err := os.Create("invoice.pdf")
	if err != nil {
		log.Println("os.Create", err)
	}
	defer fe.Close()
	_, err = fe.Write(pdfContent)
	if err != nil {
		log.Println("os.Write", err)
	}
```
response APIs
```go
    serve := http.NewServeMux()
	serve.Handle("/invoice", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/pdf")
		w.Write(pdfContent)
	}))
    http.ListenAndServe(":3000", pdfContent)
```