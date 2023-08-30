package app

import (
	"fmt"
	"image"

	_ "image/jpeg"
	_ "image/png"

	"github.com/mokiat/lacking/util/resource"
	_ "golang.org/x/image/bmp"
)

func openImage(locator resource.ReadLocator, path string) (image.Image, error) {
	in, err := locator.ReadResource(path)
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
