package app

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func invertFile(input, output string) error {
	file, err := os.Open(input)
	if err != nil {
		return fmt.Errorf("open image: %w", err)
	}
	source, format, decodeErr := image.Decode(file)
	closeErr := file.Close()
	if decodeErr != nil {
		return fmt.Errorf("decode PNG or JPEG: %w", decodeErr)
	}
	if closeErr != nil {
		return fmt.Errorf("close image: %w", closeErr)
	}
	if output == "" {
		output = input
	} else {
		switch strings.ToLower(filepath.Ext(output)) {
		case ".png":
			format = "png"
		case ".jpg", ".jpeg":
			format = "jpeg"
		default:
			return fmt.Errorf("output must have a .png, .jpg, or .jpeg extension")
		}
	}
	result := invert(source)
	return writeImage(output, func(w io.Writer) error {
		if format == "png" {
			return png.Encode(w, result)
		}
		// JPEG has no alpha channel. Flatten transparent pixels onto white.
		opaque := image.NewRGBA(result.Bounds())
		draw.Draw(opaque, opaque.Bounds(), image.White, image.Point{}, draw.Src)
		draw.Draw(opaque, opaque.Bounds(), result, result.Bounds().Min, draw.Over)
		return jpeg.Encode(w, opaque, &jpeg.Options{Quality: 95})
	})
}

func invert(source image.Image) image.Image {
	bounds := source.Bounds()
	switch source.ColorModel() {
	case color.Gray16Model, color.RGBA64Model, color.NRGBA64Model:
		result := image.NewNRGBA64(bounds)
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				c := color.NRGBA64Model.Convert(source.At(x, y)).(color.NRGBA64)
				result.SetNRGBA64(x, y, color.NRGBA64{R: ^c.R, G: ^c.G, B: ^c.B, A: c.A})
			}
		}
		return result
	default:
		result := image.NewNRGBA(bounds)
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				c := color.NRGBAModel.Convert(source.At(x, y)).(color.NRGBA)
				result.SetNRGBA(x, y, color.NRGBA{R: ^c.R, G: ^c.G, B: ^c.B, A: c.A})
			}
		}
		return result
	}
}

func writeImage(path string, encode func(io.Writer) error) error {
	mode := os.FileMode(0644)
	if info, err := os.Stat(path); err == nil {
		if !info.Mode().IsRegular() {
			return fmt.Errorf("output is not a regular file: %s", path)
		}
		mode = info.Mode().Perm()
		// Follow existing symlinks rather than replacing the link itself.
		path, err = filepath.EvalSymlinks(path)
		if err != nil {
			return fmt.Errorf("resolve output: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("stat output: %w", err)
	}
	temp, err := os.CreateTemp(filepath.Dir(path), ".invert-image-*")
	if err != nil {
		return fmt.Errorf("create output: %w", err)
	}
	defer os.Remove(temp.Name())
	if err := encode(temp); err != nil {
		temp.Close()
		return fmt.Errorf("encode image: %w", err)
	}
	if err := temp.Chmod(mode); err != nil {
		temp.Close()
		return fmt.Errorf("set output permissions: %w", err)
	}
	if err := temp.Close(); err != nil {
		return fmt.Errorf("close output: %w", err)
	}
	if err := os.Rename(temp.Name(), path); err != nil {
		return fmt.Errorf("replace output: %w", err)
	}
	return nil
}
