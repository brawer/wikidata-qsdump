// SPDX-FileCopyrightText: 2023 Sascha Brawer <sascha@brawer.ch>
// SPDX-License-Identifier: MIT

package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"gitlab.com/tozd/go/errors"
	"gitlab.com/tozd/go/mediawiki"
)

func findEntitiesDump(dumpsPath string) (time.Time, string, error) {
	path := filepath.Join(dumpsPath, "wikidatawiki", "entities", "latest-all.json.bz2")
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return time.Time{}, "", err
	}

	parts := strings.Split(resolved, string(os.PathSeparator))
	date, err := time.Parse("20060102", parts[len(parts)-2])
	if err != nil {
		return time.Time{}, "", err
	}

	// The symlink can change any time on the file system, such as
	// when Wikimedia generates a new dump right between the call
	// to EvalSymlinks() and our client opening the returned path.
	// To avoid this race condition, we need to return the resolved
	// path here, not the symlink.
	return date, resolved, nil
}

func extractQuickStatements(dumpPath string, writer io.Writer) error {
	var mutex sync.Mutex
	err := mediawiki.ProcessWikidataDump(
		context.Background(),
		&mediawiki.ProcessDumpConfig{
			Path: dumpPath,
		},
		func(_ context.Context, e mediawiki.Entity) errors.E {
			var buf bytes.Buffer
			buf.Grow(65535)
			if err := writeEntity(&e, &buf); err != nil {
				return err
			}

			mutex.Lock()
			defer mutex.Unlock()
			if _, err := writer.Write(buf.Bytes()); err != nil {
				return errors.WithStack(err)
			}
			return nil
		})
	if err != nil {
		return err
	}

	return nil
}

func writeEntity(e *mediawiki.Entity, w io.Writer) errors.E {
	if err := writeLangValues(e.ID, 'L', &e.Labels, w); err != nil {
		return err
	}
	if err := writeLangValues(e.ID, 'D', &e.Descriptions, w); err != nil {
		return err
	}
	if err := writeAliases(e, w); err != nil {
		return err
	}
	if err := writeClaims(e, w); err != nil {
		return err
	}
	if err := writeSiteLinks(e, w); err != nil {
		return err
	}

	return nil
}

func writeClaims(e *mediawiki.Entity, out io.Writer) errors.E {
	type claim struct {
		property   uint64
		statements []mediawiki.Statement
	}
	claims := make([]claim, 0, len(e.Claims))
	for key, statements := range e.Claims {
		if key[0] == 'P' {
			property, err := strconv.ParseUint(key[1:len(key)], 10, 64)
			if err != nil {
				return errors.WithStack(err)
			}
			claims = append(claims, claim{property, statements})
		}
	}
	sort.Slice(claims, func(i, j int) bool {
		return claims[i].property < claims[j].property
	})
	for _, claim := range claims {
		sort.Slice(claim.statements, func(i, j int) bool {
			return claim.statements[i].Rank < claim.statements[j].Rank
		})
		var pidbuf strings.Builder
		pidbuf.WriteRune('P')
		pidbuf.WriteString(strconv.FormatUint(claim.property, 10))
		pid := pidbuf.String()
		for _, statement := range claim.statements {
			if err := writeStatement(e.ID, pid, &statement, out); err != nil {
				return err
			}
		}
	}

	return nil
}

func writeStatement(qid string, pid string, stmt *mediawiki.Statement, out io.Writer) errors.E {
	var line bytes.Buffer
	line.WriteString(qid)
	line.WriteRune('\t')
	line.WriteString(pid)
	line.WriteRune('\t')
	switch stmt.Rank {
	case mediawiki.Preferred:
		line.WriteRune('↑')
	case mediawiki.Deprecated:
		line.WriteRune('↓')
	}

	if ok, err := writeSnak(&stmt.MainSnak, &line); !ok {
		return err
	}

	for _, qualifier := range stmt.QualifiersOrder {
		for _, val := range stmt.Qualifiers[qualifier] {
			if val.DataType != nil && val.DataValue != nil {
				var qualBuf bytes.Buffer
				qualBuf.WriteRune('\t')
				qualBuf.WriteString(qualifier)
				qualBuf.WriteRune('\t')
				ok, err := writeValue(&val, &qualBuf)
				if err != nil {
					return err
				}
				if ok {
					if _, err := line.Write(qualBuf.Bytes()); err != nil {
						return errors.WithStack(err)
					}
				}
			}
		}
	}

	for i, ref := range stmt.References {
		writeReference(&ref, i == 0, &line)
	}

	line.WriteRune('\n')

	if _, err := out.Write(line.Bytes()); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func writeReference(ref *mediawiki.Reference, first bool, out *bytes.Buffer) errors.E {
	var buf bytes.Buffer
	for i, key := range ref.SnaksOrder {
		for j, snak := range ref.Snaks[key] {
			buf.WriteRune('\t')
			if i == 0 && j == 0 && !first {
				buf.WriteRune('!')
			}
			buf.WriteRune('S')
			buf.WriteString(snak.Property[1:len(snak.Property)])
			buf.WriteRune('\t')
			if ok, err := writeSnak(&snak, &buf); !ok {
				return err
			}
		}
	}
	out.Write(buf.Bytes())
	return nil
}

func writeSnak(snak *mediawiki.Snak, out *bytes.Buffer) (bool, errors.E) {
	switch snak.SnakType {
	case mediawiki.Value:
		return writeValue(snak, out)

	case mediawiki.SomeValue:
		out.WriteString("somevalue")
		return true, nil

	case mediawiki.NoValue:
		out.WriteString("novalue")
		return true, nil
	}

	return false, nil
}

func writeValue(snak *mediawiki.Snak, out *bytes.Buffer) (bool, errors.E) {
	switch *snak.DataType {
	case mediawiki.WikiBaseItem:
		if val, ok := snak.DataValue.Value.(mediawiki.WikiBaseEntityIDValue); ok {
			out.WriteString(val.ID)
		} else {
			return false, nil
		}

	case mediawiki.ExternalID:
		fallthrough

	case mediawiki.CommonsMedia:
		fallthrough

	case mediawiki.URL:
		fallthrough

	case mediawiki.String:
		if val, ok := snak.DataValue.Value.(mediawiki.StringValue); ok {
			if err := writeQuotedString(string(val), out); err != nil {
				return false, err
			}
		} else {
			return false, nil
		}

	case mediawiki.Quantity:
		if val, ok := snak.DataValue.Value.(mediawiki.QuantityValue); ok {
			writeQuantity(&val, out)
		} else {
			return false, nil
		}

	case mediawiki.Time:
		if val, ok := snak.DataValue.Value.(mediawiki.TimeValue); ok {
			if val.Time.Year() > 0 {
				out.WriteRune('+')
			}
			out.WriteString(val.Time.Format(time.RFC3339))
			out.WriteRune('/')
			out.WriteString(strconv.FormatUint(uint64(val.Precision), 10))
			if val.Calendar == mediawiki.Julian {
				out.WriteString("/J")
			}
		} else {
			return false, nil
		}

	case mediawiki.GlobeCoordinate:
		if val, ok := snak.DataValue.Value.(mediawiki.GlobeCoordinateValue); ok {
			out.WriteRune('@')
			out.WriteString(strconv.FormatFloat(val.Latitude, 'f', 7, 32))
			out.WriteRune('/')
			out.WriteString(strconv.FormatFloat(val.Longitude, 'f', 7, 32))
		} else {
			return false, nil
		}

	case mediawiki.MonolingualText:
		if val, ok := snak.DataValue.Value.(mediawiki.MonolingualTextValue); ok {
			out.WriteString(val.Language)
			out.WriteRune(':')
			if err := writeQuotedString(val.Text, out); err != nil {
				return false, err
			}
		} else {
			return false, nil
		}

	default:
		return false, nil
	}

	return true, nil
}

func writeQuantity(val *mediawiki.QuantityValue, out *bytes.Buffer) {
	out.WriteString(val.Amount.String())
	if val.LowerBound != nil && val.UpperBound != nil {
		out.WriteRune('[')
		out.WriteString(val.LowerBound.String())
		out.WriteRune(',')
		out.WriteString(val.UpperBound.String())
		out.WriteRune(']')
	}

	const prefix = "http://www.wikidata.org/entity/Q"
	if strings.HasPrefix(val.Unit, prefix) {
		out.WriteRune('U')
		out.WriteString(val.Unit[len(prefix):len(val.Unit)])
	}
}

func writeQuotedString(s string, out *bytes.Buffer) errors.E {
	if _, err := out.WriteRune('"'); err != nil {
		return errors.WithStack(err)
	}
	for _, c := range s {
		if c < 0x20 || c == '"' || c == '\\' {
			out.WriteString(fmt.Sprintf("\\u%04x", uint64(c)))
		} else {
			if _, err := out.WriteRune(c); err != nil {
				return errors.WithStack(err)
			}
		}
	}
	if _, err := out.WriteRune('"'); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func writeLangValues(qid string, code rune, vals *map[string]mediawiki.LanguageValue, out io.Writer) errors.E {
	langvals := make([]mediawiki.LanguageValue, 0, len(*vals))
	for _, v := range *vals {
		langvals = append(langvals, v)
	}
	sort.Slice(langvals, func(i, j int) bool {
		return langvals[i].Language < langvals[j].Language
	})
	for _, v := range langvals {
		var line bytes.Buffer
		line.WriteString(qid)
		line.WriteRune('\t')
		line.WriteRune(code)
		line.WriteString(v.Language)
		line.WriteRune('\t')
		if err := writeQuotedString(v.Value, &line); err != nil {
			return err
		}
		line.WriteRune('\n')
		if _, err := out.Write(line.Bytes()); err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

func writeAliases(e *mediawiki.Entity, out io.Writer) errors.E {
	langs := make([]string, 0, len(e.Aliases))
	for lang, _ := range e.Aliases {
		langs = append(langs, lang)
	}
	sort.Strings(langs)
	for _, lang := range langs {
		for _, alias := range e.Aliases[lang] {
			var line bytes.Buffer
			line.WriteString(e.ID)
			line.WriteString("\tA")
			line.WriteString(lang)
			line.WriteRune('\t')
			if err := writeQuotedString(alias.Value, &line); err != nil {
				continue
			}
			line.WriteRune('\n')
			if _, err := out.Write(line.Bytes()); err != nil {
				return errors.WithStack(err)
			}
		}
	}
	return nil
}

func writeSiteLinks(e *mediawiki.Entity, out io.Writer) errors.E {
	type link struct {
		site  string
		title string
	}
	links := make([]link, 0, len(e.SiteLinks))
	for _, val := range e.SiteLinks {
		links = append(links, link{val.Site, val.Title})
	}
	sort.Slice(links, func(i, j int) bool {
		return links[i].site < links[j].site
	})
	for _, link := range links {
		var line bytes.Buffer
		line.WriteString(e.ID)
		line.WriteString("\tS")
		line.WriteString(link.site)
		line.WriteRune('\t')
		if err := writeQuotedString(link.title, &line); err != nil {
			return err
		}
		line.WriteRune('\n')
		if _, err := out.Write(line.Bytes()); err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}
