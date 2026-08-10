package pdfengine

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"net/http"
	"strings"
	"time"

	"fratelli-feccia/internal/dto"
)

// Claude vision fallback for zones with no text layer (scanned PDFs): the
// zone is cropped out of the rendered page and sent to the Anthropic API for
// verbatim transcription. Only active when an API key is configured.

const (
	cropPaddingPx  = 6
	minCropPx      = 8
	maxAspectRatio = 4.0
)

type subImager interface {
	SubImage(image.Rectangle) image.Image
}

// cropZone cuts the zone out of the rendered page; tall-and-narrow crops are
// rotated 90° so the vision model reads them horizontally.
func cropZone(img image.Image, f dto.PdfTemplateFieldDTO) (pngBytes []byte, rotated bool) {
	bounds := img.Bounds()
	iw, ih := bounds.Dx(), bounds.Dy()
	x := int(f.X * float64(iw))
	y := int(f.Y * float64(ih))
	w := int(f.W * float64(iw))
	h := int(f.H * float64(ih))
	if w < minCropPx || h < minCropPx {
		return nil, false
	}
	x0 := max(0, x-cropPaddingPx)
	y0 := max(0, y-cropPaddingPx)
	x1 := min(iw, x+w+cropPaddingPx)
	y1 := min(ih, y+h+cropPaddingPx)

	si, ok := img.(subImager)
	if !ok {
		return nil, false
	}
	crop := si.SubImage(image.Rect(bounds.Min.X+x0, bounds.Min.Y+y0, bounds.Min.X+x1, bounds.Min.Y+y1))

	cw, ch := crop.Bounds().Dx(), crop.Bounds().Dy()
	if cw > 0 && float64(ch)/float64(cw) > maxAspectRatio {
		crop = rotate90(crop)
		rotated = true
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, crop); err != nil {
		return nil, false
	}
	return buf.Bytes(), rotated
}

// rotate90 rotates the image 90° counter-clockwise (vertical text becomes readable).
func rotate90(src image.Image) image.Image {
	b := src.Bounds()
	dst := image.NewRGBA(image.Rect(0, 0, b.Dy(), b.Dx()))
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			dst.Set(y-b.Min.Y, b.Max.X-1-x, src.At(x, y))
		}
	}
	return dst
}

const visionSentinel = "EMPTY_NO_TEXT"

var visionHTTP = &http.Client{Timeout: 60 * time.Second}

// claudeVision transcribes the text physically present in the crop via the
// Anthropic API (plain HTTP, no SDK).
func claudeVision(ctx context.Context, apiKey string, pngBytes []byte, rotated bool) (string, error) {
	rotationHint := ""
	if rotated {
		rotationHint = " The image was rotated 90 degrees to normalize vertical text - read accordingly."
	}
	prompt := fmt.Sprintf(
		"This image is a cropped region from a business document.%s\n"+
			"Your task: transcribe ONLY the text that is physically printed/written inside this image.\n"+
			"Rules:\n"+
			"- If you can see clear alphanumeric text, return it verbatim.\n"+
			"- If the region is blank or contains only graphical elements, return exactly: %s\n"+
			"- Never guess or infer text from context.\n"+
			"- No explanations, no quotes - just the raw text or %s.",
		rotationHint, visionSentinel, visionSentinel)

	payload := map[string]any{
		"model":      "claude-haiku-4-5-20251001",
		"max_tokens": 256,
		"messages": []map[string]any{{
			"role": "user",
			"content": []map[string]any{
				{
					"type": "image",
					"source": map[string]string{
						"type":       "base64",
						"media_type": "image/png",
						"data":       base64.StdEncoding.EncodeToString(pngBytes),
					},
				},
				{"type": "text", "text": prompt},
			},
		}},
	}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := visionHTTP.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var out struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	if out.Error != nil {
		return "", fmt.Errorf("anthropic: %s", out.Error.Message)
	}
	if len(out.Content) == 0 {
		return "", nil
	}
	text := strings.TrimSpace(out.Content[0].Text)
	if text == visionSentinel {
		return "", nil
	}
	return text, nil
}
