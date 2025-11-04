// Copyright 2015 <chaishushan{AT}gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tiff

import (
	"fmt"
	"image"
)

func newImageWithIFD(r image.Rectangle, ifd *IFD) (m image.Image, err error) {
	switch ifd.ImageType() {
	case ImageType_Bilevel, ImageType_BilevelInvert:
		m = image.NewGray(r)
	case ImageType_Gray, ImageType_GrayInvert:
		if ifd.Depth() == 16 {
			m = image.NewGray16(r)
		} else {
			m = image.NewGray(r)
		}
	case ImageType_Paletted:
		m = image.NewPaletted(r, ifd.ColorMap())
	case ImageType_NRGBA:
		if ifd.Depth() == 16 {
			m = image.NewNRGBA64(r)
		} else {
			m = image.NewNRGBA(r)
		}
	case ImageType_RGB, ImageType_RGBA:
		if ifd.Depth() == 16 {
			m = image.NewRGBA64(r)
		} else {
			m = image.NewRGBA(r)
		}
	case ImageType_YCbCr:
		compression := ifd.Compression()

		// When reading from a JPEG compression stream, we directly use the
		// JPEG data, the destination image needs to be RGBA to be able to draw
		// the original image onto it. YCbCr doesn't have a Set method since
		// it does not have addressable pixels due to the subsampling.
		if compression == TagValue_CompressionType_JPEG {
			m = image.NewRGBA(r)
			return
		}

		subsampleRatio, ok := ifd.TagGetter().GetYCbCrSubSampling()
		if !ok {
			return nil, fmt.Errorf("YCbCrSubSampling not found in tags")
		}

		if len(subsampleRatio) != 2 {
			return nil, fmt.Errorf("YCbCrSubSampling subsampling length must be 2")
		}

		var ratio image.YCbCrSubsampleRatio
		if subsampleRatio[0] == 4 && subsampleRatio[1] == 4 {
			ratio = image.YCbCrSubsampleRatio444
		} else if subsampleRatio[0] == 2 && subsampleRatio[1] == 2 {
			ratio = image.YCbCrSubsampleRatio422
		} else if subsampleRatio[0] == 2 && subsampleRatio[1] == 0 {
			ratio = image.YCbCrSubsampleRatio420
		} else if subsampleRatio[0] == 4 && subsampleRatio[1] == 0 {
			ratio = image.YCbCrSubsampleRatio440
		} else if subsampleRatio[0] == 1 && subsampleRatio[1] == 1 {
			ratio = image.YCbCrSubsampleRatio411
		} else if subsampleRatio[0] == 1 && subsampleRatio[1] == 0 {
			ratio = image.YCbCrSubsampleRatio410
		} else {
			err = fmt.Errorf("tiff: Decode, unknown YCbCr Subsample Ratio")
			return
		}
		m = image.NewYCbCr(r, ratio)
	case ImageType_CMYK:
		m = image.NewCMYK(r)
	}
	if m == nil {
		err = fmt.Errorf("tiff: Decode, unknown format")
		return
	}
	return
}
