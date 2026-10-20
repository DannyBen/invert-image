package app

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestCLI(t *testing.T) {
	for _, tc := range []struct {
		args    []string
		want    string
		wantErr bool
	}{
		{args: []string{"--help"}, want: usage},
		{args: []string{"-h"}, want: usage},
		{args: []string{"--version"}, want: "invert-image 1.2.3\n"},
		{wantErr: true},
		{args: []string{"one", "two", "three"}, wantErr: true},
		{args: []string{"missing.png"}, wantErr: true},
		{args: []string{"input.png", ""}, wantErr: true},
	} {
		var out bytes.Buffer
		err := Run(tc.args, "1.2.3", &out)
		if (err != nil) != tc.wantErr || out.String() != tc.want {
			t.Errorf("Run(%q) = %q, %v; want %q, error=%t", tc.args, out.String(), err, tc.want, tc.wantErr)
		}
	}
}

func TestPNGOutputAndInPlace(t *testing.T) {
	source := image.NewNRGBA(image.Rect(0, 0, 3, 1))
	pixels := []color.NRGBA{{R: 20, G: 80, B: 160, A: 128}, {R: 0, G: 127, B: 255, A: 255}, {R: 50, G: 100, B: 150, A: 0}}
	for x, c := range pixels {
		source.SetNRGBA(x, 0, c)
	}
	input := filepath.Join(t.TempDir(), "input.png")
	original := encodePNG(t, input, source)
	output := filepath.Join(filepath.Dir(input), "output.PNG")
	if err := Run([]string{input, output}, "", io.Discard); err != nil {
		t.Fatal(err)
	}
	got, format := readImage(t, output)
	if format != "png" {
		t.Fatalf("format = %q", format)
	}
	for x, c := range pixels {
		want := color.NRGBA{R: ^c.R, G: ^c.G, B: ^c.B, A: c.A}
		if actual := color.NRGBAModel.Convert(got.At(x, 0)); actual != want {
			t.Errorf("pixel %d = %v, want %v", x, actual, want)
		}
	}
	if after, _ := os.ReadFile(input); !bytes.Equal(after, original) {
		t.Fatal("separate output changed input")
	}
	for i := 0; i < 2; i++ {
		if err := Run([]string{input}, "", io.Discard); err != nil {
			t.Fatal(err)
		}
	}
	restored, _ := readImage(t, input)
	for x, want := range pixels {
		if got := color.NRGBAModel.Convert(restored.At(x, 0)); got != want {
			t.Errorf("twice-inverted pixel %d = %v, want %v", x, got, want)
		}
	}
	assertNoTempFiles(t, filepath.Dir(input))
}

func TestPNG16Bit(t *testing.T) {
	source := image.NewNRGBA64(image.Rect(0, 0, 1, 1))
	source.SetNRGBA64(0, 0, color.NRGBA64{R: 1234, G: 23456, B: 45678, A: 32768})
	path := filepath.Join(t.TempDir(), "deep.png")
	encodePNG(t, path, source)
	if err := Run([]string{path, path}, "", io.Discard); err != nil {
		t.Fatal(err)
	}
	got, _ := readImage(t, path)
	want := color.NRGBA64{R: 65535 - 1234, G: 65535 - 23456, B: 65535 - 45678, A: 32768}
	if actual := color.NRGBA64Model.Convert(got.At(0, 0)); actual != want {
		t.Fatalf("pixel = %v, want %v", actual, want)
	}
}

func TestJPEGInPlaceAndConversion(t *testing.T) {
	source := image.NewGray(image.Rect(0, 0, 8, 8))
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			source.SetGray(x, y, color.Gray{Y: 40})
		}
	}
	var data bytes.Buffer
	if err := jpeg.Encode(&data, source, &jpeg.Options{Quality: 100}); err != nil {
		t.Fatal(err)
	}
	input := filepath.Join(t.TempDir(), "input.jpg")
	if err := os.WriteFile(input, data.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	if err := Run([]string{input}, "", io.Discard); err != nil {
		t.Fatal(err)
	}
	got, format := readImage(t, input)
	if format != "jpeg" {
		t.Fatalf("format = %q", format)
	}
	c := color.NRGBAModel.Convert(got.At(0, 0)).(color.NRGBA)
	if c.R < 213 || c.R > 217 || c.G != c.R || c.B != c.R {
		t.Fatalf("JPEG pixel = %v, want approximately gray 215", c)
	}
	output := filepath.Join(filepath.Dir(input), "output.png")
	if err := Run([]string{input, output}, "", io.Discard); err != nil {
		t.Fatal(err)
	}
	_, format = readImage(t, output)
	if format != "png" {
		t.Fatalf("converted format = %q", format)
	}
}

func TestPNGToJPEGFlattensTransparency(t *testing.T) {
	source := image.NewNRGBA(image.Rect(0, 0, 8, 8))
	input := filepath.Join(t.TempDir(), "transparent.png")
	encodePNG(t, input, source)
	output := filepath.Join(filepath.Dir(input), "output.jpeg")
	if err := Run([]string{input, output}, "", io.Discard); err != nil {
		t.Fatal(err)
	}
	got, format := readImage(t, output)
	if c := color.NRGBAModel.Convert(got.At(0, 0)).(color.NRGBA); format != "jpeg" || c.R < 254 || c.G < 254 || c.B < 254 {
		t.Fatalf("format=%s pixel=%v, want white JPEG", format, c)
	}
}

func TestFailuresPreserveFiles(t *testing.T) {
	dir := t.TempDir()
	invalid := filepath.Join(dir, "invalid.png")
	original := []byte("not an image")
	if err := os.WriteFile(invalid, original, 0600); err != nil {
		t.Fatal(err)
	}
	if err := Run([]string{invalid}, "", io.Discard); err == nil {
		t.Fatal("expected decode error")
	}
	if got, _ := os.ReadFile(invalid); !bytes.Equal(got, original) {
		t.Fatal("decode failure damaged original")
	}
	valid := filepath.Join(dir, "valid.png")
	encodePNG(t, valid, image.NewGray(image.Rect(0, 0, 1, 1)))
	for _, output := range []string{filepath.Join(dir, "bad.gif"), filepath.Join(dir, "missing", "out.png"), dir} {
		if err := Run([]string{valid, output}, "", io.Discard); err == nil {
			t.Errorf("expected error for output %q", output)
		}
	}
	destination := filepath.Join(dir, "destination.png")
	if err := os.WriteFile(destination, original, 0600); err != nil {
		t.Fatal(err)
	}
	failure := errors.New("encoder failed")
	err := writeImage(destination, func(w io.Writer) error {
		if _, err := w.Write([]byte("partial output")); err != nil {
			t.Fatal(err)
		}
		return failure
	})
	if !errors.Is(err, failure) {
		t.Fatalf("error = %v, want encoder failure", err)
	}
	if got, _ := os.ReadFile(destination); !bytes.Equal(got, original) {
		t.Fatal("encode failure damaged destination")
	}
	assertNoTempFiles(t, dir)
}

func TestInvertNonzeroBoundsAndPalette(t *testing.T) {
	source := image.NewPaletted(image.Rect(3, 4, 5, 6), color.Palette{color.NRGBA{R: 12, G: 34, B: 56, A: 255}})
	got := invert(source)
	if got.Bounds() != source.Bounds() {
		t.Fatalf("bounds = %v, want %v", got.Bounds(), source.Bounds())
	}
	if c := color.NRGBAModel.Convert(got.At(3, 4)); c != (color.NRGBA{R: 243, G: 221, B: 199, A: 255}) {
		t.Fatalf("pixel = %v", c)
	}
}

func encodePNG(t *testing.T, path string, source image.Image) []byte {
	t.Helper()
	var data bytes.Buffer
	if err := png.Encode(&data, source); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	return data.Bytes()
}

func readImage(t *testing.T, path string) (image.Image, string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	img, format, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	return img, format
}

func assertNoTempFiles(t *testing.T, dir string) {
	t.Helper()
	matches, err := filepath.Glob(filepath.Join(dir, ".invert-image-*"))
	if err != nil || len(matches) != 0 {
		t.Fatalf("leftover temporary files: %v (%v)", matches, err)
	}
}
