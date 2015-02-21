// Copyright 2014 <chaishushan{AT}gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tiff

import (
	"fmt"
	"time"
)

func (p *IFD) TagGetter() TagGetter {
	return nil
}

func (p *IFD) TagSetter() TagSetter {
	return nil
}

func (p *IFD) ImageSize() (width, height int) {
	if tag, ok := p.EntryMap[TagType_ImageWidth]; ok {
		if v := tag.GetInts(); len(v) == 1 {
			width = int(v[0])
		}
	}
	if tag, ok := p.EntryMap[TagType_ImageLength]; ok {
		if v := tag.GetInts(); len(v) == 1 {
			height = int(v[0])
		}
	}
	return
}

func (p *IFD) BitsPerSample() int {
	if tag, ok := p.EntryMap[TagType_BitsPerSample]; ok {
		v := tag.GetInts()
		for i := 1; i < len(v); i++ {
			if v[i] != v[0] {
				return 0
			}
		}
		if len(v) > 0 {
			return int(v[0])
		}
	}
	return 0
}

func (p *IFD) SamplesPerPixel() int {
	if tag, ok := p.EntryMap[TagType_SamplesPerPixel]; ok {
		if v := tag.GetInts(); len(v) == 1 {
			return int(v[0])
		}
	}
	return 0
}

func (p *IFD) PhotometricInterpretation() TagValue_PhotometricType {
	if tag, ok := p.EntryMap[TagType_PhotometricInterpretation]; ok {
		if v := tag.GetInts(); len(v) == 1 {
			return TagValue_PhotometricType(v[0])
		}
	}
	return TagValue_PhotometricType_WhiteIsZero
}

func (p *IFD) Compression() CompressType {
	if tag, ok := p.EntryMap[TagType_Compression]; ok {
		if v := tag.GetInts(); len(v) == 1 {
			return CompressType(v[0])
		}
	}
	return CompressType_Nil
}

func (p *IFD) ResolutionUnit() TagValue_ResolutionUnitType {
	if tag, ok := p.EntryMap[TagType_ResolutionUnit]; ok {
		if v := tag.GetInts(); len(v) == 1 {
			return TagValue_ResolutionUnitType(v[0])
		}
	}
	return TagValue_ResolutionUnitType_None
}

func (p *IFD) XYResolution() (x, y [2]int64) {
	if tag, ok := p.EntryMap[TagType_XResolution]; ok {
		if v := tag.GetRationals(); len(v) == 1 {
			x = v[0]
		}
	}
	if tag, ok := p.EntryMap[TagType_YResolution]; ok {
		if v := tag.GetRationals(); len(v) == 1 {
			y = v[0]
		}
	}
	return
}

func (p *IFD) ColorMap() [][3]uint16 {
	if tag, ok := p.EntryMap[TagType_ColorMap]; ok {
		v := tag.GetInts()
		if len(v) == 0 || len(v)%3 != 0 {
			return nil
		}
		colorMap := make([][3]uint16, len(v)/3)
		for i := 0; i < len(v); i += 3 {
			colorMap[i/3] = [3]uint16{
				uint16(v[i+0]),
				uint16(v[i+1]),
				uint16(v[i+2]),
			}
		}
	}
	return nil
}

func (p *IFD) CellSize() (width, height float64) {
	if tag, ok := p.EntryMap[TagType_CellWidth]; ok {
		if v := tag.GetFloats(); len(v) == 1 {
			width = float64(v[0])
		}
	}
	if tag, ok := p.EntryMap[TagType_CellLenght]; ok {
		if v := tag.GetFloats(); len(v) == 1 {
			height = float64(v[0])
		}
	}
	return
}

func (p *IFD) BlockSize() (width, height int) {
	if _, ok := p.EntryMap[TagType_TileWidth]; ok {
		if tag, ok := p.EntryMap[TagType_TileWidth]; ok {
			if v := tag.GetInts(); len(v) == 1 {
				width = int(v[0])
			}
		}
		if tag, ok := p.EntryMap[TagType_TileLength]; ok {
			if v := tag.GetInts(); len(v) == 1 {
				height = int(v[0])
			}
		}

	} else {
		if tag, ok := p.EntryMap[TagType_ImageWidth]; ok {
			if v := tag.GetInts(); len(v) == 1 {
				width = int(v[0])
			}
		}
		if tag, ok := p.EntryMap[TagType_RowsPerStrip]; ok {
			if v := tag.GetInts(); len(v) == 1 {
				height = int(v[0])
			}
		}
		if height == 0 {
			if tag, ok := p.EntryMap[TagType_ImageLength]; ok {
				if v := tag.GetInts(); len(v) == 1 {
					height = int(v[0])
				}
			}
		}
	}
	return
}

func (p *IFD) BlockOffsets() []int64 {
	if _, ok := p.EntryMap[TagType_TileWidth]; ok {
		if tag, ok := p.EntryMap[TagType_TileOffsets]; ok {
			return tag.GetInts()
		}
	} else {
		if tag, ok := p.EntryMap[TagType_StripOffsets]; ok {
			return tag.GetInts()
		}
	}
	return nil
}

func (p *IFD) BlockCounts() []int64 {
	if _, ok := p.EntryMap[TagType_TileWidth]; ok {
		if tag, ok := p.EntryMap[TagType_TileByteCounts]; ok {
			return tag.GetInts()
		}
	} else {
		if tag, ok := p.EntryMap[TagType_StripByteCounts]; ok {
			return tag.GetInts()
		}
	}
	return nil
}

func (p *IFD) Predictor() TagValue_PredictorType {
	if tag, ok := p.EntryMap[TagType_Predictor]; ok {
		if v := tag.GetInts(); len(v) == 1 {
			return TagValue_PredictorType(v[0])
		}
	}
	return TagValue_PredictorType_None
}

func (p *IFD) DocumentName() string {
	if tag, ok := p.EntryMap[TagType_DocumentName]; ok {
		return tag.GetString()
	}
	return ""
}

func (p *IFD) ImageDescription() string {
	if tag, ok := p.EntryMap[TagType_ImageDescription]; ok {
		return tag.GetString()
	}
	return ""
}

func (p *IFD) Make() string {
	if tag, ok := p.EntryMap[TagType_Make]; ok {
		return tag.GetString()
	}
	return ""
}

func (p *IFD) Model() string {
	if tag, ok := p.EntryMap[TagType_Model]; ok {
		return tag.GetString()
	}
	return ""
}

func (p *IFD) Software() string {
	if tag, ok := p.EntryMap[TagType_Software]; ok {
		return tag.GetString()
	}
	return ""
}
func (p *IFD) Artist() string {
	if tag, ok := p.EntryMap[TagType_Artist]; ok {
		return tag.GetString()
	}
	return ""
}
func (p *IFD) HostComputer() string {
	if tag, ok := p.EntryMap[TagType_HostComputer]; ok {
		return tag.GetString()
	}
	return ""
}

func (p *IFD) DateTime() (t time.Time) {
	if tag, ok := p.EntryMap[TagType_DateTime]; ok {
		var year, month, day, hour, min, sec int
		if _, err := fmt.Sscanf(tag.GetString(), "%d:%d:%d %d:%d:%d",
			&year, &month, &day,
			&hour, &min, &sec,
		); err == nil {
			t = time.Date(year, time.Month(month), day, hour, min, sec, 0, nil)
		}
	}
	return
}
func (p *IFD) Copyright() string {
	if tag, ok := p.EntryMap[TagType_Copyright]; ok {
		return tag.GetString()
	}
	return ""
}
