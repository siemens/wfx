// SPDX-FileCopyrightText: 2016 Tom Hudson
//
// SPDX-License-Identifier: MIT
//
// SPDX-FileNotice: Origin: https://github.com/tomnomnom/gron

package gron

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"

	"github.com/Southclaws/fault"
)

type token struct {
	text string
	typ  tokenType
}

type tokenType int

const (
	typBare tokenType = iota
	typNumericKey
	typQuotedKey
	typDot
	typLBrace
	typRBrace
	typEquals
	typSemi
	typString
	typNumber
	typTrue
	typFalse
	typNull
	typEmptyArray
	typEmptyObject
)

type statement []token

func (s statement) String() string {
	out := make([]string, 0, len(s))
	for _, t := range s {
		out = append(out, t.format())
	}
	return strings.Join(out, "")
}

func (t token) format() string {
	if t.typ == typEquals {
		return " " + t.text + " "
	}
	return t.text
}

func (s statement) withBare(key string) statement {
	result := make(statement, len(s), len(s)+2)
	copy(result, s)
	return append(result, token{".", typDot}, token{key, typBare})
}

func (s statement) withQuotedKey(key string) statement {
	result := make(statement, len(s), len(s)+3)
	copy(result, s)
	return append(result, token{"[", typLBrace}, token{quoteString(key), typQuotedKey}, token{"]", typRBrace})
}

func (s statement) withNumericKey(index int) statement {
	result := make(statement, len(s), len(s)+3)
	copy(result, s)
	return append(result, token{"[", typLBrace}, token{strconv.Itoa(index), typNumericKey}, token{"]", typRBrace})
}

type statements []statement

func (ss *statements) addWithValue(path statement, value token) {
	result := make(statement, len(path), len(path)+3)
	copy(result, path)
	result = append(result, token{"=", typEquals}, value, token{";", typSemi})
	*ss = append(*ss, result)
}

func (ss statements) Len() int      { return len(ss) }
func (ss statements) Swap(i, j int) { ss[i], ss[j] = ss[j], ss[i] }

func (ss statements) Less(a, b int) bool {
	diffIndex := -1
	for i := range ss[a] {
		if len(ss[b]) < i+1 {
			return false
		}
		if ss[a][i] == ss[b][i] {
			continue
		}
		diffIndex = i
		break
	}
	if diffIndex == -1 {
		return true
	}

	left, right := ss[a][diffIndex], ss[b][diffIndex]
	if left.typ == typEquals {
		return true
	}
	if right.typ == typEquals {
		return false
	}
	if left.typ == typNumericKey && right.typ == typNumericKey {
		leftIndex, _ := strconv.Atoi(left.text)
		rightIndex, _ := strconv.Atoi(right.text)
		return leftIndex < rightIndex
	}
	if left.typ != typNumber || right.typ != typNumber {
		return left.text < right.text
	}
	leftNumber, _ := json.Number(left.text).Float64()
	rightNumber, _ := json.Number(right.text).Float64()
	return leftNumber < rightNumber
}

func statementsFromJSON(r io.Reader, prefix statement) (statements, error) {
	var value any
	decoder := json.NewDecoder(r)
	decoder.UseNumber()
	if err := decoder.Decode(&value); err != nil {
		return nil, fault.Wrap(err)
	}
	result := make(statements, 0, 32)
	result.fill(prefix, value)
	return result, nil
}

func (ss *statements) fill(prefix statement, value any) {
	ss.addWithValue(prefix, valueToken(value))
	switch v := value.(type) {
	case map[string]any:
		keys := make([]string, 0, len(v))
		for key := range v {
			keys = append(keys, key)
		}
		for _, key := range keys {
			if validIdentifier(key) {
				ss.fill(prefix.withBare(key), v[key])
			} else {
				ss.fill(prefix.withQuotedKey(key), v[key])
			}
		}
	case []any:
		for index, item := range v {
			ss.fill(prefix.withNumericKey(index), item)
		}
	}
}

func valueToken(value any) token {
	switch v := value.(type) {
	case map[string]any:
		return token{"{}", typEmptyObject}
	case []any:
		return token{"[]", typEmptyArray}
	case json.Number:
		return token{v.String(), typNumber}
	case string:
		return token{quoteString(v), typString}
	case bool:
		if v {
			return token{"true", typTrue}
		}
		return token{"false", typFalse}
	case nil:
		return token{"null", typNull}
	default:
		return token{"", typNull}
	}
}

var reservedWords = map[string]bool{
	"break": true, "case": true, "catch": true, "class": true,
	"const": true, "continue": true, "debugger": true, "default": true,
	"delete": true, "do": true, "else": true, "export": true, "extends": true,
	"false": true, "finally": true, "for": true, "function": true, "if": true,
	"import": true, "in": true, "instanceof": true, "new": true, "null": true,
	"return": true, "super": true, "switch": true, "this": true, "throw": true,
	"true": true, "try": true, "typeof": true, "var": true, "void": true,
	"while": true, "with": true, "yield": true,
}

func validIdentifier(s string) bool {
	if reservedWords[s] || s == "" {
		return false
	}
	for i, r := range s {
		if i == 0 && !validFirstRune(r) || i != 0 && !validSecondaryRune(r) {
			return false
		}
	}
	return true
}

func validFirstRune(r rune) bool {
	return unicode.In(
		r,
		unicode.Lu,
		unicode.Ll,
		unicode.Lm,
		unicode.Lo,
		unicode.Nl,
	) || r == '$' || r == '_'
}

func validSecondaryRune(r rune) bool {
	return validFirstRune(r) || unicode.In(r, unicode.Mn, unicode.Mc, unicode.Nd, unicode.Pc)
}

func quoteString(s string) string {
	out := &bytes.Buffer{}
	_ = out.WriteByte('"')
	for _, r := range s {
		switch r {
		case '\\':
			_, _ = out.WriteString(`\\`)
		case '"':
			_, _ = out.WriteString(`\"`)
		case '\b':
			_, _ = out.WriteString(`\b`)
		case '\f':
			_, _ = out.WriteString(`\f`)
		case '\n':
			_, _ = out.WriteString(`\n`)
		case '\r':
			_, _ = out.WriteString(`\r`)
		case '\t':
			_, _ = out.WriteString(`\t`)
		case '\u2028':
			_, _ = out.WriteString(`\u2028`)
		case '\u2029':
			_, _ = out.WriteString(`\u2029`)
		default:
			if unicode.IsControl(r) {
				_, _ = fmt.Fprintf(out, `\u%04X`, r)
			} else {
				_, _ = out.WriteRune(r)
			}
		}
	}
	_ = out.WriteByte('"')
	return out.String()
}
