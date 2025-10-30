// Copyright 2014 <chaishushan{AT}gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tiff

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"io"
)

func (p *IFD) BlocksAcross() int {
	imageWidth, _ := p.TagGetter().GetImageWidth()
	if imageWidth == 0 {
		return 0
	}
	blockWidth, _ := p.TagGetter().GetTileWidth()
	if blockWidth > 0 {
		return int((imageWidth + blockWidth - 1) / blockWidth)
	}
	return 1
}

func (p *IFD) BlocksDown() int {
	imageHeight, _ := p.TagGetter().GetImageLength()
	if imageHeight == 0 {
		return 0
	}
	blockHeight, _ := p.TagGetter().GetTileLength()
	if blockHeight > 0 {
		return int((imageHeight + blockHeight - 1) / blockHeight)
	}
	blockHeight, _ = p.TagGetter().GetRowsPerStrip()
	if blockHeight == 0 || blockHeight > imageHeight {
		blockHeight = imageHeight
	}
	return int((imageHeight + blockHeight - 1) / blockHeight)
}

func (p *IFD) BlockBounds(col, row int) image.Rectangle {
	blocksAcross, blocksDown := p.BlocksAcross(), p.BlocksDown()
	if col < 0 || row < 0 || col >= blocksAcross || row >= blocksDown {
		return image.Rectangle{}
	}

	if _, ok := p.TagGetter().GetTileWidth(); ok {
		blockWidth, _ := p.TagGetter().GetTileWidth()
		blockHeight, _ := p.TagGetter().GetTileLength()

		blkW := blockWidth
		blkH := blockHeight

		xmin := col * int(blockWidth)
		ymin := row * int(blockHeight)
		xmax := xmin + int(blkW)
		ymax := ymin + int(blkH)

		return image.Rect(xmin, ymin, xmax, ymax)

	} else {
		imageWidth, _ := p.TagGetter().GetImageWidth()
		imageHeight, _ := p.TagGetter().GetImageLength()

		blockWidth := imageWidth
		blockHeight, ok := p.TagGetter().GetRowsPerStrip()
		if !ok || blockHeight == 0 {
			blockHeight = imageHeight
		}

		blkW := blockWidth
		blkH := blockHeight
		if row == blocksDown-1 && imageHeight%blockHeight != 0 {
			blkH = imageHeight % blockHeight
		}

		xmin := col * int(blockWidth)
		ymin := row * int(blockHeight)
		xmax := xmin + int(blkW)
		ymax := ymin + int(blkH)

		return image.Rect(xmin, ymin, xmax, ymax)
	}
}

func (p *IFD) BlockOffset(col, row int) int64 {
	blocksAcross, blocksDown := p.BlocksAcross(), p.BlocksDown()
	if col < 0 || row < 0 || col >= blocksAcross || row >= blocksDown {
		return 0
	}
	if _, ok := p.TagGetter().GetTileWidth(); ok {
		offsets, ok := p.TagGetter().GetTileOffsets()
		if !ok || len(offsets) != blocksAcross*blocksDown {
			return 0
		}
		return offsets[row*blocksAcross+col]
	} else {
		offsets, ok := p.TagGetter().GetStripOffsets()
		if !ok || len(offsets) != blocksAcross*blocksDown {
			return 0
		}
		return offsets[row*blocksAcross+col]
	}
}

func (p *IFD) BlockCount(col, row int) int64 {
	blocksAcross, blocksDown := p.BlocksAcross(), p.BlocksDown()
	if col < 0 || row < 0 || col >= blocksAcross || row >= blocksDown {
		return 0
	}
	if _, ok := p.TagGetter().GetTileWidth(); ok {
		counts, ok := p.TagGetter().GetTileByteCounts()
		if !ok || len(counts) != blocksAcross*blocksDown {
			return 0
		}
		return counts[row*blocksAcross+col]
	} else {
		counts, ok := p.TagGetter().GetStripByteCounts()
		if !ok || len(counts) != blocksAcross*blocksDown {
			return 0
		}
		return counts[row*blocksAcross+col]
	}
}

func (p *IFD) DecodeBlock(r io.ReadSeeker, col, row int, dst image.Image) (err error) {
	blocksAcross, blocksDown := p.BlocksAcross(), p.BlocksDown()
	if col < 0 || row < 0 || col >= blocksAcross || row >= blocksDown {
		err = fmt.Errorf("tiff: IFD.DecodeBlock, bad col/row = %d/%d", col, row)
		return
	}

	bounds := p.BlockBounds(col, row)
	offset := p.BlockOffset(col, row)
	count := p.BlockCount(col, row)

	if _, err = r.Seek(offset, 0); err != nil {
		return
	}
	limitReader := io.LimitReader(r, count)

	var data []byte
	var img image.Image
	if data, img, err = p.Compression().Decode(limitReader, bounds.Dx(), bounds.Dy(), p); err != nil {
		return
	}

	// If an image is provided by decompression, copy its content onto the dst image.
	if img != nil {
		drawImage, ok := dst.(draw.Image)
		if ok {
			tileSize := img.Bounds().Size()
			draw.Draw(drawImage, image.Rect(bounds.Min.X, bounds.Min.Y, bounds.Min.X+tileSize.X, bounds.Min.Y+tileSize.Y), img, img.Bounds().Min, draw.Src)
		}
		return
	}

	predictor, ok := p.TagGetter().GetPredictor()
	if ok && predictor == TagValue_PredictorType_Horizontal {
		if data, err = p.decodePredictor(data, bounds); err != nil {
			return
		}
	}

	err = p.decodeBlock(data, dst, bounds)
	return
}

func (p *IFD) decodePredictor(data []byte, r image.Rectangle) (out []byte, err error) {
	bpp := p.Depth()
	spp := p.Channels()

	switch bpp {
	case 16:
		var off int
		for y := r.Min.Y; y < r.Max.Y; y++ {
			off += spp * 2
			for x := 0; x < (r.Dx()-1)*spp*2; x += 2 {
				if off+2 > len(data) {
					err = fmt.Errorf("tiff: IFD.decodePredictor, not enough pixel data")
					return
				}
				v0 := p.Header.ByteOrder.Uint16(data[off-spp*2 : off-spp*2+2])
				v1 := p.Header.ByteOrder.Uint16(data[off : off+2])
				p.Header.ByteOrder.PutUint16(data[off:off+2], v1+v0)
				off += 2
			}
		}
	case 8:
		var off int
		for y := r.Min.Y; y < r.Max.Y; y++ {
			off += spp
			for x := 0; x < (r.Dx()-1)*spp; x++ {
				if off >= len(data) {
					err = fmt.Errorf("tiff: IFD.decodePredictor, not enough pixel data")
					return
				}
				data[off] += data[off-spp]
				off++
			}
		}
	default:
		err = fmt.Errorf("tiff: IFD.decodePredictor, bad BitsPerSample = %d", bpp)
		return
	}
	out = data
	return
}

func (p *IFD) decodeBlock(buf []byte, dst image.Image, r image.Rectangle) (err error) {
	xmin, ymin := r.Min.X, r.Min.Y
	xmax, ymax := r.Max.X, r.Max.Y

	rMaxX := minInt(xmax, dst.Bounds().Max.X)
	rMaxY := minInt(ymax, dst.Bounds().Max.Y)

	b := p.Bounds()
	rMaxX = minInt(rMaxX, b.Max.X)
	rMaxY = minInt(rMaxY, b.Max.Y)

	switch p.ImageType() {
	case ImageType_Gray, ImageType_GrayInvert, ImageType_Bilevel, ImageType_BilevelInvert:
		if x, bpp := p.Compression(), p.Depth(); bpp == 1 && (x == TagValue_CompressionType_G3 || x == TagValue_CompressionType_G4) {
			img := dst.(*image.Gray)
			for y := ymin; y < rMaxY; y++ {
				min := img.PixOffset(xmin, y)
				max := img.PixOffset(rMaxX, y)
				off := (y - ymin) * (xmax - xmin) * 1
				for i := min; i < max; i++ {
					img.Pix[i+0] = buf[off+0]
					off++
				}
			}
			return
		}

		if p.Depth() == 16 {
			var off int
			img := dst.(*image.Gray16)
			for y := ymin; y < rMaxY; y++ {
				for x := xmin; x < rMaxX; x++ {
					if off+2 > len(buf) {
						err = fmt.Errorf("tiff: IFD.decodeBlock, not enough pixel data")
						return
					}
					v := p.Header.ByteOrder.Uint16(buf[off : off+2])
					off += 2
					if p.ImageType() == ImageType_GrayInvert {
						v = 0xffff - v
					}
					img.SetGray16(x, y, color.Gray16{v})
				}
			}
		} else {
			bpp := uint(p.Depth())
			bitReader := newBitsReader(buf)
			img := dst.(*image.Gray)

			max := uint32((1 << uint(p.Depth())) - 1)
			for y := ymin; y < rMaxY; y++ {
				for x := xmin; x < rMaxX; x++ {
					v, ok := bitReader.ReadBits(bpp)
					if !ok {
						err = fmt.Errorf("tiff: IFD.decodeBlock, not enough pixel data")
						return
					}
					v = v * 0xff / max
					if p.ImageType() == ImageType_GrayInvert {
						v = 0xff - v
					}
					img.SetGray(x, y, color.Gray{uint8(v)})
				}
				bitReader.flushBits()
			}
		}
	case ImageType_Paletted:
		bpp := uint(p.Depth())
		bitReader := newBitsReader(buf)
		img := dst.(*image.Paletted)
		for y := ymin; y < rMaxY; y++ {
			for x := xmin; x < rMaxX; x++ {
				v, ok := bitReader.ReadBits(bpp)
				if !ok {
					err = fmt.Errorf("tiff: IFD.decodeBlock, not enough pixel data")
					return
				}
				img.SetColorIndex(x, y, uint8(v))
			}
			bitReader.flushBits()
		}
	case ImageType_RGB:
		if p.Depth() == 16 {
			var off int
			img := dst.(*image.RGBA64)
			for y := ymin; y < rMaxY; y++ {
				for x := xmin; x < rMaxX; x++ {
					if off+6 > len(buf) {
						err = fmt.Errorf("tiff: IFD.decodeBlock, not enough pixel data")
						return
					}
					r := p.Header.ByteOrder.Uint16(buf[off+0 : off+2])
					g := p.Header.ByteOrder.Uint16(buf[off+2 : off+4])
					b := p.Header.ByteOrder.Uint16(buf[off+4 : off+6])
					off += 6
					img.SetRGBA64(x, y, color.RGBA64{r, g, b, 0xffff})
				}
			}
		} else {
			img := dst.(*image.RGBA)
			for y := ymin; y < rMaxY; y++ {
				min := img.PixOffset(xmin, y)
				max := img.PixOffset(rMaxX, y)
				off := (y - ymin) * (xmax - xmin) * 3
				for i := min; i < max; i += 4 {
					if off+3 > len(buf) {
						err = fmt.Errorf("tiff: IFD.decodeBlock, not enough pixel data")
						return
					}
					img.Pix[i+0] = buf[off+0]
					img.Pix[i+1] = buf[off+1]
					img.Pix[i+2] = buf[off+2]
					img.Pix[i+3] = 0xff
					off += 3
				}
			}
		}
	case ImageType_NRGBA:
		if p.Depth() == 16 {
			var off int
			img := dst.(*image.NRGBA64)
			for y := ymin; y < rMaxY; y++ {
				for x := xmin; x < rMaxX; x++ {
					if off+8 > len(buf) {
						err = fmt.Errorf("tiff: IFD.decodeBlock, not enough pixel data")
						return
					}
					r := p.Header.ByteOrder.Uint16(buf[off+0 : off+2])
					g := p.Header.ByteOrder.Uint16(buf[off+2 : off+4])
					b := p.Header.ByteOrder.Uint16(buf[off+4 : off+6])
					a := p.Header.ByteOrder.Uint16(buf[off+6 : off+8])
					off += 8
					img.SetNRGBA64(x, y, color.NRGBA64{r, g, b, a})
				}
			}
		} else {
			img := dst.(*image.NRGBA)
			for y := ymin; y < rMaxY; y++ {
				min := img.PixOffset(xmin, y)
				max := img.PixOffset(rMaxX, y)
				i0, i1 := (y-ymin)*(xmax-xmin)*4, (y-ymin+1)*(xmax-xmin)*4
				if i1 > len(buf) {
					err = fmt.Errorf("tiff: IFD.decodeBlock, not enough pixel data")
					return
				}
				copy(img.Pix[min:max], buf[i0:i1])
			}
		}
	case ImageType_RGBA:
		if p.Depth() == 16 {
			var off int
			img := dst.(*image.RGBA64)
			for y := ymin; y < rMaxY; y++ {
				for x := xmin; x < rMaxX; x++ {
					if off+8 > len(buf) {
						err = fmt.Errorf("tiff: IFD.decodeBlock, not enough pixel data")
						return
					}
					r := p.Header.ByteOrder.Uint16(buf[off+0 : off+2])
					g := p.Header.ByteOrder.Uint16(buf[off+2 : off+4])
					b := p.Header.ByteOrder.Uint16(buf[off+4 : off+6])
					a := p.Header.ByteOrder.Uint16(buf[off+6 : off+8])
					off += 8
					img.SetRGBA64(x, y, color.RGBA64{r, g, b, a})
				}
			}
		} else {
			img := dst.(*image.RGBA)
			for y := ymin; y < rMaxY; y++ {
				min := img.PixOffset(xmin, y)
				max := img.PixOffset(rMaxX, y)
				i0, i1 := (y-ymin)*(xmax-xmin)*4, (y-ymin+1)*(xmax-xmin)*4
				if i1 > len(buf) {
					err = fmt.Errorf("tiff: IFD.decodeBlock, not enough pixel data")
					return
				}
				copy(img.Pix[min:max], buf[i0:i1])
			}
		}
	default:
		err = fmt.Errorf("tiff: IFD.decodeBlock, unknown imageType: %v", p.ImageType())
		return
	}

	return
}

func (p *IFD) EncodeBlock(w io.Writer, col, row int, dst *MemPImage) (err error) {
	return
}
