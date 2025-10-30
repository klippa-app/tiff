// Copyright 2015 <chaishushan{AT}gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tiff

import (
	"bufio"
	"bytes"
	"compress/zlib"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"io/ioutil"

	"github.com/chai2010/tiff/internal/fax"
)

func (p TagValue_CompressionType) Decode(r io.Reader, width, height int, ifd *IFD) (data []byte, img image.Image, err error) {
	switch p {
	case TagValue_CompressionType_None, TagValue_CompressionType_Nil:
		return p.decode_None(r)
	case TagValue_CompressionType_CCITT:
		return p.decode_CCITT(r)
	case TagValue_CompressionType_G3:
		return p.decode_G3(r, width, height)
	case TagValue_CompressionType_G4:
		return p.decode_G4(r, width, height)
	case TagValue_CompressionType_LZW:
		return p.decode_LZW(r)
	case TagValue_CompressionType_JPEGOld:
		return p.decode_JPEGOld(r)
	case TagValue_CompressionType_JPEG:
		return p.decode_JPEG(r, ifd)
	case TagValue_CompressionType_Deflate:
		return p.decode_Deflate(r)
	case TagValue_CompressionType_PackBits:
		return p.decode_PackBits(r)
	case TagValue_CompressionType_DeflateOld:
		return p.decode_DeflateOld(r)
	}
	err = fmt.Errorf("tiff: unsupport %v compression type", int(p))
	return
}

func (p TagValue_CompressionType) decode_None(r io.Reader) (data []byte, img image.Image, err error) {
	data, err = ioutil.ReadAll(r)
	return
}

func (p TagValue_CompressionType) decode_CCITT(r io.Reader) (data []byte, img image.Image, err error) {
	err = fmt.Errorf("tiff: unsupport %v compression type", "CCITT")
	return
}

func (p TagValue_CompressionType) decode_G3(r io.Reader, width, height int) (data []byte, img image.Image, err error) {
	return p.decode_G4(r, width, height)
}

func (p TagValue_CompressionType) decode_G4(r io.Reader, width, height int) (data []byte, img image.Image, err error) {
	br, ok := r.(io.ByteReader)
	if !ok {
		br = bufio.NewReader(r)
	}

	return fax.DecodeG4Pixels(br, width, height)
}

func (p TagValue_CompressionType) decode_LZW(r io.Reader) (data []byte, img image.Image, err error) {
	lzwReader := newLzwReader(r, lzwMSB, 8)
	data, err = ioutil.ReadAll(lzwReader)
	lzwReader.Close()
	return
}

func (p TagValue_CompressionType) decode_JPEGOld(r io.Reader) (data []byte, img image.Image, err error) {
	println("decode_JPEGOld")
	err = fmt.Errorf("tiff: unsupport %v compression type", "JPEGOld")
	return
}

func (p TagValue_CompressionType) decode_JPEG(r io.Reader, ifd *IFD) (data []byte, img image.Image, err error) {
	var decodedImage image.Image
	var imageReader io.Reader

	// To decode the JPEG data, we need the Huffman and Quantization table.
	// Tiff adds this to the header so that it can share this over multiple
	// image slices.
	jpegTables, ok := ifd.TagGetter().GetJPEGTables()

	// Merge image data and jpeg tables.
	if ok && jpegTables != nil && len(jpegTables) > 4 {
		var imageData []byte
		imageData, err = io.ReadAll(r)
		if err != nil {
			return
		}

		newImage := bytes.NewBuffer(nil)

		// First verify some stuff before merging.
		if jpegTables[0] != 0xff || jpegTables[1] != 0xd8 {
			err = errors.New("tiff: invalid jpeg table, does not begin with SOI marker")
			return
		}

		if jpegTables[len(jpegTables)-2] != 0xff || jpegTables[len(jpegTables)-1] != 0xd9 {
			err = errors.New("tiff: invalid jpeg table, does not end wth EOI marker")
			return
		}

		if imageData[0] != 0xff || imageData[1] != 0xd8 {
			err = errors.New("tiff: invalid image data, does not begin with SOI marker")
			return
		}

		if imageData[len(imageData)-2] != 0xff || imageData[len(imageData)-1] != 0xd9 {
			err = errors.New("tiff: invalid image data, does not end wth EOI marker")
			return
		}

		// Write the JPEG tables without the EOI marker.
		newImage.Write(jpegTables[0 : len(jpegTables)-2])

		// Write the image data without the SOI marker.
		newImage.Write(imageData[2:])

		imageReader = newImage
	} else {
		// Just try to decode the original image.
		imageReader = r
	}

	decodedImage, err = jpeg.Decode(imageReader)
	if err != nil {
		err = fmt.Errorf("tiff: could not decode JPEG image: %w", err)
		return
	}

	return nil, decodedImage, nil
}

func (p TagValue_CompressionType) decode_Deflate(r io.Reader) (data []byte, img image.Image, err error) {
	zlibReader, err := zlib.NewReader(r)
	if err != nil {
		return nil, nil, err
	}
	data, err = ioutil.ReadAll(zlibReader)
	zlibReader.Close()
	return
}

func (p TagValue_CompressionType) decode_DeflateOld(r io.Reader) (data []byte, img image.Image, err error) {
	zlibReader, err := zlib.NewReader(r)
	if err != nil {
		return nil, nil, err
	}
	data, err = ioutil.ReadAll(zlibReader)
	zlibReader.Close()
	return
}

func (p TagValue_CompressionType) decode_PackBits(r io.Reader) (data []byte, img image.Image, err error) {
	type byteReader interface {
		io.Reader
		io.ByteReader
	}

	buf := make([]byte, 128)
	dst := make([]byte, 0, 1024)
	br, ok := r.(byteReader)
	if !ok {
		br = bufio.NewReader(r)
	}

	for {
		b, err := br.ReadByte()
		if err != nil {
			if err == io.EOF {
				return dst, nil, nil
			}
			return nil, nil, err
		}
		code := int(int8(b))
		switch {
		case code >= 0:
			n, err := io.ReadFull(br, buf[:code+1])
			if err != nil {
				return nil, nil, err
			}
			dst = append(dst, buf[:n]...)
		case code == -128:
			// No-op.
		default:
			if b, err = br.ReadByte(); err != nil {
				return nil, nil, err
			}
			for j := 0; j < 1-code; j++ {
				buf[j] = b
			}
			dst = append(dst, buf[:1-code]...)
		}
	}
}
