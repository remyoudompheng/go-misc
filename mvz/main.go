package main

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	//"image/png"
	"io"
	//"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

func main() {
	file := os.Args[1]
	dir := os.Args[2]

	f, err := os.Open(file)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	m := readHeader(f)
	log.Printf("%+v", m.Header)

	os.MkdirAll(dir, 0755)

	var buf []byte
	readJpeg := func(n int) (image.Image, error) {
		if len(buf) < n {
			buf = make([]byte, n)
		}
		_, err := io.ReadFull(f, buf[:n])
		if err != nil {
			return nil, err
		}
		return jpeg.Decode(bytes.NewReader(buf))
	}
	var bufsize image.Rectangle
	bufsize.Max.X = int(m.Width)
	bufsize.Max.Y = int(m.Height)
	imgBuf := image.NewRGBA(bufsize)
	for page, tsizes := range m.ImgSizes {
		log.Printf("reading page %d", page+1)
		// read tiles: they are arranged in columns
		for i, tsize := range tsizes[:m.NX*m.NY] {
			x, y := i/int(m.NY), i%int(m.NY)
			off, _ := f.Seek(0, os.SEEK_CUR)
			img, err := readJpeg(int(tsize))
			if err != nil {
				log.Printf("cannot read tile %d(%d,%d) at 0x%x: %s",
					page+1, x, y, off, err)
			} else {
				var dest image.Rectangle
				dest.Min.X = 256 * (i / int(m.NY))
				dest.Min.Y = 256 * (i % int(m.NY))
				dest.Max = bufsize.Max
				draw.Src.Draw(imgBuf, dest, img, img.Bounds().Min)
			}
		}
		// skip full page image
		off, _ := f.Seek(0, os.SEEK_CUR)
		_, err := readJpeg(int(tsizes[len(tsizes)-2]))
		if err != nil {
			log.Printf("cannot read page view %d at 0x%x: %s",
				page+1, off, err)
		}

		outpath := filepath.Join(dir, fmt.Sprintf("page%04d.jpg", page+1))
		w, err := os.Create(outpath)
		if err != nil {
			log.Fatal(err)
		}
		err = jpeg.Encode(w, imgBuf, &jpeg.Options{Quality: 85})
		if err != nil {
			log.Fatal(err)
		}
		w.Close()
		log.Printf("written image to %s", outpath)

		// clear buffer
		for i := range imgBuf.Pix {
			imgBuf.Pix[i] = 0
		}
	}

	for page, tsizes := range m.ImgSizes {
		// skip thumbnail.
		off, _ := f.Seek(0, os.SEEK_CUR)
		_, err := readJpeg(int(tsizes[len(tsizes)-1]))
		if err != nil {
			log.Printf("thumbnail of page %d (offset 0x%x) is corrupted: %s",
				page+1, off, err)
		}
	}
	off, _ := f.Seek(0, os.SEEK_CUR)
	stat, _ := f.Stat()
	if off < stat.Size() {
		log.Printf("WARNING: trailing bytes: %d", stat.Size()-off)
	}
}
