package utils

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/png"
	"io"

	"github.com/disintegration/imaging"
)

func ResizeImage(r io.Reader, width, height int) ([]byte, error) {
	src, _, err := image.Decode(r)
	if err != nil {
		return nil, err
	}

	dst := imaging.Fill(src, width, height, imaging.Center, imaging.Lanczos)

	var buf bytes.Buffer
	err = png.Encode(&buf, dst)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type ImageType string

const (
	ImageTypePNG ImageType = "image/png"
	ImageTypeSVG ImageType = "image/svg+xml"
)

func EncodeToBase64(imgType ImageType, imgData []byte) string {
	return "data:" + string(imgType) + ";base64," + base64.StdEncoding.EncodeToString(imgData)
}
