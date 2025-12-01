/*
 * MIT License
 *
 * Copyright (c) 2025 Nicolas JUHEL
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 *
 *
 */

package perm

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
)

func parseString(s string) (Perm, error) {
	s = strings.TrimSpace(s)
	s = strings.Replace(s, "\"", "", -1) // nolint
	s = strings.Replace(s, "'", "", -1)  // nolint

	if v, e := strconv.ParseUint(s, 8, 32); e != nil {
		return parseLetterString(s)
	} else if v > math.MaxUint32 {
		return Perm(0), fmt.Errorf("invalid permission")
	} else {
		return Perm(v), nil
	}
}

func parseLetterString(s string) (Perm, error) {
	s = strings.TrimSpace(s)
	s = strings.Replace(s, "\"", "", -1) // nolint
	s = strings.Replace(s, "'", "", -1)  // nolint

	if len(s) != 9 && len(s) != 10 {
		return 0, fmt.Errorf("invalid permission")
	}

	var perm os.FileMode = 0
	startIdx := uint8(0)

	// if file type is given, then use it
	if len(s) == 10 {
		switch s[0] {
		case '-': // Fichier régulier
			perm |= 0
		case 'd': // Répertoire
			perm |= os.ModeDir
		case 'l': // Lien symbolique
			perm |= os.ModeSymlink
		case 'c': // Périphérique de caractères
			perm |= os.ModeDevice | os.ModeCharDevice
		case 'b': // Périphérique de blocs
			perm |= os.ModeDevice
		case 'p': // FIFO (tube nommé)
			perm |= os.ModeNamedPipe
		case 's': // Socket
			perm |= os.ModeSocket
		case 'D': // Porte (Door)
			perm |= os.ModeIrregular
		default:
			return 0, fmt.Errorf("invalid file type character: %c", s[0])
		}
		startIdx = 1
	}

	// Fonction pour convertir un groupe de 3 caractères en valeur octale
	convertGroup := func(chars string) (uint8, error) {
		if len(chars) != 3 {
			return 0, fmt.Errorf("invalid permission group length")
		}

		var value uint8 = 0
		if chars[0] == 'r' {
			value += 4
		} else if chars[0] != '-' {
			return 0, fmt.Errorf("invalid read permission character: %c", chars[0])
		}

		if chars[1] == 'w' {
			value += 2
		} else if chars[1] != '-' {
			return 0, fmt.Errorf("invalid write permission character: %c", chars[1])
		}

		if chars[2] == 'x' {
			value += 1
		} else if chars[2] != '-' {
			return 0, fmt.Errorf("invalid execute permission character: %c", chars[2])
		}

		return value, nil
	}

	// convert each group of 3 chars
	for i := uint8(0); i < 3; i++ {
		start := startIdx + i*3
		end := start + 3
		if int(end) > len(s) {
			return 0, fmt.Errorf("invalid permission string format")
		}

		group := s[start:end]
		value, err := convertGroup(group)
		if err != nil {
			return 0, err
		}

		// Shift by 6, 3, or 0 bits depending on the group (owner, group, others)
		// Accumulate permissions for each group
		perm |= os.FileMode(value) << uint(6-i*3)
	}

	return Perm(perm), nil
}

func (p *Perm) parseString(s string) error {
	if v, e := parseString(s); e != nil {
		return e
	} else {
		*p = v
		return nil
	}
}

func (p *Perm) unmarshall(val []byte) error {
	if tmp, err := ParseByte(val); err != nil {
		return err
	} else {
		*p = tmp
		return nil
	}
}
