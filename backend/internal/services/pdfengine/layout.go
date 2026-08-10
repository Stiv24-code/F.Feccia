package pdfengine

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"image"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"fratelli-feccia/internal/dto"
)

// Poppler side of the engine: page rendering (pdftoppm) and the text layer
// with word coordinates (pdftotext -bbox-layout), both in PDF points.

const renderDPI = 108 // 1.5 × 72pt, come OrderMesh

type bboxWord struct {
	XMin float64 `xml:"xMin,attr"`
	YMin float64 `xml:"yMin,attr"`
	XMax float64 `xml:"xMax,attr"`
	YMax float64 `xml:"yMax,attr"`
	Text string  `xml:",chardata"`
}

type bboxLine struct {
	Words []bboxWord `xml:"word"`
}

type bboxBlock struct {
	XMin  float64    `xml:"xMin,attr"`
	YMin  float64    `xml:"yMin,attr"`
	XMax  float64    `xml:"xMax,attr"`
	YMax  float64    `xml:"yMax,attr"`
	Lines []bboxLine `xml:"line"`
}

type bboxPage struct {
	Width  float64
	Height float64
	Blocks []bboxBlock
}

// flow > block: collect blocks across flows
type bboxFlow struct {
	Blocks []bboxBlock `xml:"block"`
}

type bboxPageXML struct {
	Width  float64    `xml:"width,attr"`
	Height float64    `xml:"height,attr"`
	Flows  []bboxFlow `xml:"flow"`
}

type bboxDoc struct {
	Pages []bboxPageXML `xml:"body>doc>page"`
}

// pdfTextLayout runs pdftotext -bbox-layout and returns per-page word/block
// boxes in PDF points.
func pdfTextLayout(dir, pdfPath string) ([]bboxPage, error) {
	outPath := filepath.Join(dir, "layout.xhtml")
	cmd := exec.Command("pdftotext", "-bbox-layout", pdfPath, outPath)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("pdftotext: %v — %s", err, strings.TrimSpace(stderr.String()))
	}
	raw, err := os.ReadFile(outPath)
	if err != nil {
		return nil, err
	}
	var doc bboxDoc
	if err := xml.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("parse layout: %w", err)
	}
	pages := make([]bboxPage, len(doc.Pages))
	for i, p := range doc.Pages {
		page := bboxPage{Width: p.Width, Height: p.Height}
		for _, f := range p.Flows {
			page.Blocks = append(page.Blocks, f.Blocks...)
		}
		pages[i] = page
	}
	return pages, nil
}

var pageNumRe = regexp.MustCompile(`-(\d+)\.png$`)

// renderAllPages renders every page to PNG at renderDPI and returns the file
// paths ordered by page number.
func renderAllPages(dir, pdfPath string) ([]string, error) {
	prefix := filepath.Join(dir, "page")
	cmd := exec.Command("pdftoppm", "-png", "-r", strconv.Itoa(renderDPI), pdfPath, prefix)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("pdftoppm: %v — %s", err, strings.TrimSpace(stderr.String()))
	}
	matches, err := filepath.Glob(prefix + "-*.png")
	if err != nil {
		return nil, err
	}
	sort.Slice(matches, func(i, j int) bool {
		return pageFileNum(matches[i]) < pageFileNum(matches[j])
	})
	return matches, nil
}

func pageFileNum(path string) int {
	m := pageNumRe.FindStringSubmatch(path)
	if len(m) < 2 {
		return 0
	}
	n, _ := strconv.Atoi(m[1])
	return n
}

// renderOnePage renders a single 0-based page to PNG and returns the decoded
// image — used lazily by the vision fallback.
func renderOnePage(dir, pdfPath string, pageNum int) (image.Image, error) {
	prefix := filepath.Join(dir, fmt.Sprintf("crop-%d", pageNum))
	p := strconv.Itoa(pageNum + 1)
	cmd := exec.Command("pdftoppm", "-png", "-r", strconv.Itoa(renderDPI), "-f", p, "-l", p, pdfPath, prefix)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("pdftoppm pagina %s: %v — %s", p, err, strings.TrimSpace(stderr.String()))
	}
	matches, _ := filepath.Glob(prefix + "-*.png")
	if len(matches) == 0 {
		return nil, fmt.Errorf("pagina %s non renderizzata", p)
	}
	f, err := os.Open(matches[0])
	if err != nil {
		return nil, err
	}
	defer f.Close()
	img, err := png.Decode(f)
	if err != nil {
		return nil, err
	}
	return img, nil
}

func blockText(b bboxBlock) string {
	var parts []string
	for _, l := range b.Lines {
		for _, w := range l.Words {
			if t := strings.TrimSpace(w.Text); t != "" {
				parts = append(parts, t)
			}
		}
	}
	return strings.Join(parts, " ")
}

// zoneText joins the words whose center falls inside the zone rectangle,
// in document (reading) order.
func zoneText(page bboxPage, f dto.PdfTemplateFieldDTO) string {
	if page.Width == 0 || page.Height == 0 {
		return ""
	}
	x0, y0 := f.X*page.Width, f.Y*page.Height
	x1, y1 := x0+f.W*page.Width, y0+f.H*page.Height
	var parts []string
	for _, b := range page.Blocks {
		// skip blocks fully outside the zone
		if b.XMax < x0 || b.XMin > x1 || b.YMax < y0 || b.YMin > y1 {
			continue
		}
		for _, l := range b.Lines {
			for _, w := range l.Words {
				cx, cy := (w.XMin+w.XMax)/2, (w.YMin+w.YMax)/2
				if cx >= x0 && cx <= x1 && cy >= y0 && cy <= y1 {
					if t := strings.TrimSpace(w.Text); t != "" {
						parts = append(parts, t)
					}
				}
			}
		}
	}
	return strings.Join(parts, " ")
}
