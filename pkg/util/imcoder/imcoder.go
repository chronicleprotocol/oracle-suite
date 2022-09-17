package imcoder

import (
	"encoding/binary"
	"errors"
	"hash/crc32"
	"image"
	"image/color"
)

const (
	version          = 1
	metaBlocksOffset = 15
	dataHeaderOffset = 10
)

type Options struct {
	// BlockSize is the size in pixels of the side of the block. A smaller size
	// yields images with smaller dimensions, but when combined with lossy
	// compression, the resulting image may be impossible to decode.
	BlockSize uint
	// BitsPerChan is the number of bits used to represent each RGB value.
	// It must be in the range [1, 8]. A higher value yields images with
	// smaller number of blocks, but when combined with lossy compression, the
	// resulting image may be impossible to decode.
	BitsPerChan uint
}

// Encode encodes the data into an image.
func Encode(data []byte, opts Options) (image.Image, error) {
	if opts.BlockSize == 0 {
		return nil, errors.New("imcoder: invalid block size")
	}
	if opts.BitsPerChan == 0 || opts.BitsPerChan > 8 {
		return nil, errors.New("imcoder: invalid bits per channel")
	}
	data = addDataHeader(data)
	rgb := make([]uint8, rgbLengthForData(len(data), opts.BitsPerChan))
	writeMetaBlocks(rgb, opts.BitsPerChan)
	writeDataBlocks(rgb[metaBlocksOffset:], data, opts.BitsPerChan)
	return drawImage(rgb, opts.BlockSize), nil
}

// Decode decodes the data from an image.
func Decode(img image.Image) ([]byte, error) {
	blockSize := measureBlockSize(img)
	if blockSize == 0 {
		return nil, errors.New("imcoder: invalid block size")
	}
	rgbLen := rgbLengthForImage(img, blockSize)
	if rgbLen == 0 {
		return nil, errors.New("imcoder: invalid image dimensions")
	}
	rgb := make([]uint8, rgbLen)
	if len(rgb) < metaBlocksOffset {
		return nil, errors.New("imcoder: invalid image")
	}
	readImage(rgb, img, blockSize)
	bpc := readMetaBlocks(rgb)
	data := make([]uint8, dataLength(len(rgb), bpc))
	readDataBlocks(data, rgb[metaBlocksOffset:], bpc)
	return verifyData(data)
}

// addDataHeader prepends the data with a header that contains the data length
// and a CRC32 checksum.
func addDataHeader(data []byte) []byte {
	d := make([]byte, len(data)+dataHeaderOffset)
	binary.LittleEndian.PutUint16(d, uint16(version))
	binary.LittleEndian.PutUint32(d[2:], uint32(len(data)))        // data length
	binary.LittleEndian.PutUint32(d[6:], crc32.ChecksumIEEE(data)) // checksum
	copy(d[dataHeaderOffset:], data)
	return d
}

// verifyData verifies the data and returns the data without the header.
func verifyData(data []byte) ([]byte, error) {
	if len(data) < dataHeaderOffset {
		return nil, errors.New("imcoder: invalid data")
	}
	if binary.LittleEndian.Uint16(data) != version {
		return nil, errors.New("imcoder: unsupported version")
	}
	dataLen := binary.LittleEndian.Uint32(data[2:])
	checksum := binary.LittleEndian.Uint32(data[6:])
	if dataLen > uint32(len(data)-dataHeaderOffset) {
		return nil, errors.New("imcoder: invalid length")
	}
	data = data[dataHeaderOffset : dataLen+dataHeaderOffset]
	if checksum != crc32.ChecksumIEEE(data) {
		return nil, errors.New("imcoder: invalid checksum")
	}
	return data, nil
}

// writeMetaBlocks prepends an image with 5 black and white meta blocks.
// The first two blocks are alignment blocks that are used to measure the
// block size during decoding. The next 3 black and white blocks are used
// to store the number of bits per channel.
func writeMetaBlocks(rgb []uint8, bpc uint) {
	copy(rgb, []uint8{0, 0, 0, 255, 255, 255})
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			if (bpc-1)&(1<<uint(i)) != 0 {
				rgb[6+i*3+j] = 255
			}
		}
	}
}

// readMetaBlocks reads the number of bits per channel from the meta blocks.
func readMetaBlocks(rgb []uint8) uint {
	var bpc uint
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			if rgb[6+i*3+j] >= 128 {
				bpc |= 1 << uint(i)
			}
		}
	}
	return bpc + 1
}

// writeDataBlocks encodes the data into RGB values. The data is stored in the
// rgb slice. The rgb slice must have sufficient length to store the data.
func writeDataBlocks(rgb []uint8, data []byte, bpc uint) {
	mask := uint8(0xff << (8 - bpc))
	for i, j := 0, 0; i < len(data)*8; i, j = i+int(bpc), j+1 {
		q := i / 8
		r := i % 8
		c := uint8(data[q] << r)
		if q < len(data)-1 {
			c |= data[q+1] >> (8 - r)
		}
		c &= mask
		for k := uint(0); k < 8; k += bpc {
			c |= c >> bpc
		}
		rgb[j] = c
	}
}

// readDataBlocks decodes the data from RGB values. The data is stored in the
// data slice. The data slice must have sufficient length to store the data.
func readDataBlocks(data []byte, rgb []uint8, bpc uint) {
	mask := uint8(0xff << (8 - bpc))
	for i, j := 0, 0; i < len(rgb); i, j = i+1, j+1 {
		v := rgb[i] & mask
		p := j * int(bpc)
		q := p / 8
		r := p % 8
		data[q] |= v >> r
		if 8-r < int(bpc) {
			data[q+1] |= v << (8 - r)
		}
	}
}

// drawImage draws the RGB values into an image.
func drawImage(rgb []uint8, bs uint) image.Image {
	bn := 0                     // block number
	xy := imageXY(len(rgb), bs) // image width and height
	img := image.NewRGBA(image.Rect(0, 0, xy, xy))
	for y := 0; y < xy; y += int(bs) {
		for x := 0; x < xy; x += int(bs) {
			c := color.RGBA{A: 255}
			if bn < len(rgb) {
				c.R = rgb[bn]
			}
			if bn+1 < len(rgb) {
				c.G = rgb[bn+1]
			}
			if bn+2 < len(rgb) {
				c.B = rgb[bn+2]
			}
			bn += 3
			for by := 0; by < int(bs); by++ {
				for bx := 0; bx < int(bs); bx++ {
					img.Set(x+bx, y+by, c)
				}
			}
		}
	}
	return img
}

// readImage reads the RGB values from an image and stores them in the rgb
// slice. The rgb slice must have sufficient length to store the data.
func readImage(rgb []uint8, img image.Image, bs uint) {
	xy := img.Bounds().Dx() // image width and height
	ba := uint64(bs * bs)   // block area
	bn := 0                 // block number
	for y := 0; y < xy; y += int(bs) {
		for x := 0; x < xy; x += int(bs) {
			if bn+3 >= len(rgb)-1 {
				break
			}
			var sr, sg, sb uint64
			for by := 0; by < int(bs); by++ {
				for bx := 0; bx < int(bs); bx++ {
					r, g, b, _ := img.At(x+bx, y+by).RGBA()
					sr += uint64(r >> 8)
					sg += uint64(g >> 8)
					sb += uint64(b >> 8)
				}
			}
			rgb[bn] = uint8(sr / ba)
			rgb[bn+1] = uint8(sg / ba)
			rgb[bn+2] = uint8(sb / ba)
			bn += 3
		}
	}
}

// measureBlockSize measures the block size in the image using the first two
// black and white alignment blocks.
func measureBlockSize(img image.Image) uint {
	xy := img.Bounds().Dx() // image width and height
	bs := 0                 // block size
	for x := 0; x < xy; x++ {
		r, g, b, _ := img.At(x, 0).RGBA()
		if r>>8 >= 128 && g>>8 >= 128 && b>>8 >= 128 {
			bs = x
			break
		}
	}
	// The size of the side of the image must be divisible by the block size.
	// If this is not the case, the estimated block size is probably wrong due
	// to high image compression. In such a case, the following code tries to
	// find the closest divisible block size.
	for i := 0; i < bs; i++ {
		if xy%(bs+i) == 0 {
			bs += i
			break
		}
		if xy%(bs-i) == 0 {
			bs -= i
			break
		}
	}
	return uint(bs)
}

// rgbLengthForData calculates the length of the RGB slice needed to store the
// given amount of data.
func rgbLengthForData(dataLen int, bpc uint) int {
	return divRoundUp(dataLen*8, int(bpc)) + metaBlocksOffset
}

// rgbLengthForImage calculates the maximum length of the RGB slice needed to store
// the data from the given image.
func rgbLengthForImage(img image.Image, bs uint) int {
	dx, dy := img.Bounds().Dx(), img.Bounds().Dy()
	if dy < dx {
		return 0
	}
	return (dx * dx) / int(bs*bs) * 3
}

// dataLength calculates the length of the data stored in the RGB slice.
func dataLength(rgbLen int, bpc uint) int {
	return divRoundUp(rgbLen*int(bpc), 8)
}

func imageXY(rgbLen int, bs uint) int {
	xy := 1
	for xy*xy < rgbLen {
		xy++
	}
	return xy * int(bs)
}

func divRoundUp(a, b int) int {
	if a%b == 0 {
		return a / b
	}
	return a/b + 1
}
