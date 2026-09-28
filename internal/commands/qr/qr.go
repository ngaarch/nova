package qrcmd

import (
	"fmt"
	"io"
	"os"

	"nova/internal/command"
	"nova/internal/output"
)

// Code represents a generated QR code matrix.
type Code struct {
	Version int      `json:"version"`
	Size    int      `json:"size"`
	Matrix  [][]bool `json:"matrix"`
	Text    string   `json:"text"`
}

// Command returns the registered Command instance for qr.
func Command() *command.Command {
	return &command.Command{
		Name:        "qr",
		Aliases:     []string{"qrcode"},
		Summary:     "Generate and render high-resolution UTF-8 half-block QR codes in the terminal",
		Usage:       "nova qr [flags] <text>",
		Description: "Encode text, URLs, or Wi-Fi configurations into scannable terminal QR codes using Unicode half-blocks.",
		Phase:       20,
		Run:         Run,
	}
}

// Run executes the qr command.
func Run(ctx *command.Context, args []string) error {
	for _, arg := range args {
		if arg == "--plain" {
			ctx.Printer.Mode = output.ModePlain
		} else if arg == "--json" {
			ctx.Printer.Mode = output.ModeJSON
		}
	}

	opts := ParseFlags(args)

	// Read from file or stdin if requested
	if opts.File != "" {
		if opts.File == "-" {
			data, err := io.ReadAll(ctx.Stdin)
			if err != nil {
				return fmt.Errorf("read stdin: %w", err)
			}
			opts.Text = string(data)
		} else {
			data, err := os.ReadFile(opts.File)
			if err != nil {
				return fmt.Errorf("read file %s: %w", opts.File, err)
			}
			opts.Text = string(data)
		}
	}

	if opts.Text == "" {
		return fmt.Errorf("missing text to encode (usage: nova qr [flags] <text>)")
	}

	code, err := Encode(opts.Text)
	if err != nil {
		return fmt.Errorf("generate QR code: %w", err)
	}

	return Render(ctx, code, opts)
}

// GF(256) tables with primitive polynomial 0x11d
var (
	expTable [512]byte
	logTable [256]byte
	gfInit   bool
)

func initGF() {
	if gfInit {
		return
	}
	var val = 1
	for i := 0; i < 255; i++ {
		expTable[i] = byte(val)
		expTable[i+255] = byte(val)
		logTable[val] = byte(i)
		val = val << 1
		if val >= 256 {
			val = val ^ 0x11d
		}
	}
	gfInit = true
}

func gfMul(x, y byte) byte {
	if x == 0 || y == 0 {
		return 0
	}
	return expTable[int(logTable[x])+int(logTable[y])]
}

func rsGeneratorPoly(degree int) []byte {
	poly := []byte{1}
	for i := 0; i < degree; i++ {
		next := []byte{1, expTable[i]}
		res := make([]byte, len(poly)+len(next)-1)
		for j, v := range poly {
			for k, w := range next {
				res[j+k] ^= gfMul(v, w)
			}
		}
		poly = res
	}
	return poly
}

func rsEncode(data []byte, ecCount int) []byte {
	gen := rsGeneratorPoly(ecCount)
	res := make([]byte, len(data)+ecCount)
	copy(res, data)

	for i := 0; i < len(data); i++ {
		coef := res[i]
		if coef != 0 {
			for j := 0; j < len(gen); j++ {
				res[i+j] ^= gfMul(gen[j], coef)
			}
		}
	}
	return res[len(data):]
}

type qrVersionConfig struct {
	version      int
	dataCapacity int
	ecCodewords  int
}

var versionConfigs = []qrVersionConfig{
	{version: 1, dataCapacity: 16, ecCodewords: 10},
	{version: 2, dataCapacity: 28, ecCodewords: 16},
	{version: 3, dataCapacity: 44, ecCodewords: 26},
	{version: 4, dataCapacity: 64, ecCodewords: 36},
	{version: 5, dataCapacity: 86, ecCodewords: 48},
	{version: 6, dataCapacity: 108, ecCodewords: 64},
	{version: 7, dataCapacity: 130, ecCodewords: 72},
}

// Encode generates a standard QR code matrix.
func Encode(text string) (Code, error) {
	initGF()

	rawBytes := []byte(text)
	var chosenCfg *qrVersionConfig
	for _, cfg := range versionConfigs {
		// In byte mode: 4 bits mode + 8 bits length + data
		// Total payload bytes = len(rawBytes) + 2 (approx overhead)
		if len(rawBytes)+3 <= cfg.dataCapacity {
			chosenCfg = &cfg
			break
		}
	}

	if chosenCfg == nil {
		return Code{}, fmt.Errorf("text too large (%d bytes, max 127 bytes supported)", len(rawBytes))
	}

	// 1. Bitstream encoding (8-bit Byte mode)
	var bitBuf []bool
	// Mode indicator: 0100 for Byte mode
	appendBits(&bitBuf, 0x4, 4)
	// Character count indicator: 8 bits for versions 1-9
	appendBits(&bitBuf, uint32(len(rawBytes)), 8)
	// Data
	for _, b := range rawBytes {
		appendBits(&bitBuf, uint32(b), 8)
	}
	// Terminator (up to 4 zeroes)
	maxBits := chosenCfg.dataCapacity * 8
	for len(bitBuf)%8 != 0 && len(bitBuf) < maxBits {
		bitBuf = append(bitBuf, false)
	}
	// Pad bytes (0xEC, 0x11 alternating)
	dataCodewords := make([]byte, 0, chosenCfg.dataCapacity)
	for i := 0; i < len(bitBuf); i += 8 {
		var val byte
		for b := 0; b < 8; b++ {
			if i+b < len(bitBuf) && bitBuf[i+b] {
				val |= 1 << (7 - b)
			}
		}
		dataCodewords = append(dataCodewords, val)
	}

	padBytes := []byte{0xec, 0x11}
	padIdx := 0
	for len(dataCodewords) < chosenCfg.dataCapacity {
		dataCodewords = append(dataCodewords, padBytes[padIdx%2])
		padIdx++
	}

	// 2. Reed-Solomon Error Correction
	ecCodewords := rsEncode(dataCodewords, chosenCfg.ecCodewords)

	// Combine data + EC
	allCodewords := append(dataCodewords, ecCodewords...)

	// 3. Matrix construction
	size := 17 + 4*chosenCfg.version
	matrix := make([][]bool, size)
	reserved := make([][]bool, size)
	for i := range matrix {
		matrix[i] = make([]bool, size)
		reserved[i] = make([]bool, size)
	}

	// Place Finders
	placeFinder(matrix, reserved, 0, 0)
	placeFinder(matrix, reserved, size-7, 0)
	placeFinder(matrix, reserved, 0, size-7)

	// Timing patterns
	for i := 8; i < size-8; i++ {
		matrix[6][i] = (i % 2) == 0
		reserved[6][i] = true
		matrix[i][6] = (i % 2) == 0
		reserved[i][6] = true
	}

	// Alignment patterns for version >= 2
	if chosenCfg.version >= 2 {
		alignPos := 4*chosenCfg.version + 10
		placeAlignment(matrix, reserved, alignPos, alignPos)
	}

	// Dark module
	matrix[4*chosenCfg.version+9][8] = true
	reserved[4*chosenCfg.version+9][8] = true

	// Reserve format info areas around finders
	for i := 0; i < 9; i++ {
		reserved[8][i] = true
		reserved[i][8] = true
		reserved[8][size-1-i] = true
		reserved[size-1-i][8] = true
	}

	// Place data bits in zig-zag
	allBits := make([]bool, 0, len(allCodewords)*8)
	for _, cw := range allCodewords {
		for b := 7; b >= 0; b-- {
			allBits = append(allBits, (cw&(1<<b)) != 0)
		}
	}

	bitIdx := 0
	upwards := true
	for right := size - 1; right > 0; right -= 2 {
		if right == 6 {
			right-- // skip timing column 6
		}

		var rowStart, rowEnd, rowStep int
		if upwards {
			rowStart, rowEnd, rowStep = size-1, -1, -1
		} else {
			rowStart, rowEnd, rowStep = 0, size, 1
		}

		for r := rowStart; r != rowEnd; r += rowStep {
			for colOffset := 0; colOffset < 2; colOffset++ {
				c := right - colOffset
				if reserved[r][c] {
					continue
				}

				val := false
				if bitIdx < len(allBits) {
					val = allBits[bitIdx]
					bitIdx++
				}

				// Apply standard mask 0: (row + col) % 2 == 0
				if (r+c)%2 == 0 {
					val = !val
				}
				matrix[r][c] = val
			}
		}
		upwards = !upwards
	}

	// Write Format Information (Level M, Mask 0)
	// Format bits for Level M (00) and Mask 0 (000): 00000 -> with BCH 101010000010010
	formatInfo := uint16(0x5412) // standard masked format info for M, mask 0
	placeFormatInfo(matrix, formatInfo, size)

	return Code{
		Version: chosenCfg.version,
		Size:    size,
		Matrix:  matrix,
		Text:    text,
	}, nil
}

func appendBits(buf *[]bool, val uint32, count int) {
	for i := count - 1; i >= 0; i-- {
		*buf = append(*buf, (val&(1<<i)) != 0)
	}
}

func placeFinder(m, res [][]bool, r, c int) {
	for dr := -1; dr <= 7; dr++ {
		for dc := -1; dc <= 7; dc++ {
			row, col := r+dr, c+dc
			if row >= 0 && row < len(m) && col >= 0 && col < len(m) {
				res[row][col] = true
				if dr >= 0 && dr <= 6 && dc >= 0 && dc <= 6 {
					if dr == 0 || dr == 6 || dc == 0 || dc == 6 || (dr >= 2 && dr <= 4 && dc >= 2 && dc <= 4) {
						m[row][col] = true
					} else {
						m[row][col] = false
					}
				} else {
					m[row][col] = false
				}
			}
		}
	}
}

func placeAlignment(m, res [][]bool, r, c int) {
	for dr := -2; dr <= 2; dr++ {
		for dc := -2; dc <= 2; dc++ {
			row, col := r+dr, c+dc
			if row >= 0 && row < len(m) && col >= 0 && col < len(m) {
				res[row][col] = true
				if dr == -2 || dr == 2 || dc == -2 || dc == 2 || (dr == 0 && dc == 0) {
					m[row][col] = true
				} else {
					m[row][col] = false
				}
			}
		}
	}
}

func placeFormatInfo(m [][]bool, bits uint16, size int) {
	// 15 bits of format info
	for i := 0; i < 15; i++ {
		bit := (bits & (1 << i)) != 0

		// Top-left finder
		var r1, c1 int
		if i < 6 {
			r1, c1 = 8, i
		} else if i < 8 {
			r1, c1 = 8, i+1
		} else if i == 8 {
			r1, c1 = 7, 8
		} else {
			r1, c1 = 14-i, 8
		}
		m[r1][c1] = bit

		// Split across bottom-left and top-right finders
		var r2, c2 int
		if i < 8 {
			r2, c2 = size-1-i, 8
		} else {
			r2, c2 = 8, size-15+i
		}
		m[r2][c2] = bit
	}
}
