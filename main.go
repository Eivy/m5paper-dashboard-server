package main

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/chromedp/chromedp"
)

var port string

func main() {
	port = os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", handler)
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true), // headless=false に変更
		chromedp.Flag("no-sandbox", true),
	)
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()
	ctx, cancel = chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	defer cancel()
	if os.Getenv("GRAFANA_NEED_LOGIN") != "" {
		trg := os.Getenv("GRAFANA_DASHBOARD_URL")
		if trg == "" {
			trg = os.Getenv("GRAFANA_URL")
		}
		chromedp.Run(ctx,
			chromedp.Navigate(trg),
			chromedp.Sleep(2*time.Second),
			chromedp.SendKeys(`input[name='user']`, os.Getenv("GRAFANA_USER")),
			chromedp.SendKeys(`input[name='password']`, os.Getenv("GRAFANA_PASSWORD")),
			chromedp.Click(`button[type='submit']`),
		)
	}
	if err := srv.ListenAndServe(); err != nil {
		log.Print(err)
	}

}

var ctx context.Context

func handler(w http.ResponseWriter, r *http.Request) {
	var res []byte
	log.Println(path.Join(os.Getenv("GRAFANA_URL"), r.URL.Path+r.URL.RawQuery))
	trg := os.Getenv("GRAFANA_DASHBOARD_URL")
	if trg == "" {
		trg = path.Join(os.Getenv("GRAFANA_URL"), r.URL.Path+"?"+r.URL.RawQuery)
	}
	chromedp.Run(ctx,
		chromedp.Navigate(trg),
		chromedp.EmulateViewport(950, 540),
		chromedp.WaitVisible(`div.u-over`),
		chromedp.Sleep(200*time.Millisecond),
		chromedp.CaptureScreenshot(&res),
	)
	imageP, _ := png.Decode(bytes.NewReader(res))
	dst := image.NewGray(imageP.Bounds())
	for y := 0; y < imageP.Bounds().Dy(); y++ {
		for x := 0; x < imageP.Bounds().Dx(); x++ {
			c := color.GrayModel.Convert(imageP.At(x, y))
			gray, _ := c.(color.Gray)
			if gray.Y > 200 {
				dst.Set(x, y, color.Gray{Y: 255})
			} else if gray.Y < 25 {
				dst.Set(x, y, color.Gray{Y: 0})
			} else {
				dst.Set(x, y, gray)
			}
		}
	}
	tmp := filepath.Join(os.TempDir(), "grafana.jpg")
	f, err := os.Create(tmp)
	if err != nil {
		log.Fatal(err)
	}
	jpeg.Encode(f, dst, nil)
	f.Close()
	s, err := os.Stat(tmp)
	if err != nil {
		log.Fatal(err)
	}
	f, err = os.Open(tmp)
	if err != nil {
		log.Fatal(err)
	}
	w.Header().Add("Content-Length", fmt.Sprint(s.Size()))
	w.WriteHeader(http.StatusOK)
	io.Copy(w, f)
}
