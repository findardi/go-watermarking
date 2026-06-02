package codec

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
)

const jpegQuality = 90

func Decode(data []byte) (image.Image, string, error) {
	img, format, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, "", fmt.Errorf("failed to decode image: %w", err)
	}

	return img, format, nil
}

func Encode(img image.Image, format string) ([]byte, error) {
	var buf bytes.Buffer
	var err error
	switch format {
	case "jpeg":
		err = jpeg.Encode(&buf, img, &jpeg.Options{
			Quality: jpegQuality,
		})
	case "png":
		err = png.Encode(&buf, img)
	default:
		return nil, fmt.Errorf("unsupported format: %q", format)
	}

	if err != nil {
		return nil, fmt.Errorf("encode %s: %w", format, err)
	}

	return buf.Bytes(), nil
}
