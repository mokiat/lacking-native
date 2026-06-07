package app

import (
	"fmt"
	"image"

	_ "image/jpeg"
	_ "image/png"

	"github.com/mokiat/lacking/resource"
	_ "golang.org/x/image/bmp"
)

func openImage(locator resource.Locator, path string) (image.Image, error) {
	in, err := locator.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer in.Close()

	img, _, err := image.Decode(in)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}
	return img, nil
}
