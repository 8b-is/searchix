package index

import (
	"strings"
	"unicode"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/search/query"
)

type tokenKind int

const (
	tokEOF tokenKind = iota
	tokWord
	tokPhrase
	tokField
	tokLParen
	tokRParen
	tokOr
	tokAnd
	tokNot
	tokPlus
)

type queryToken struct {
	kind   tokenKind
	text   string
	field  string
	quoted bool
}

var fieldAliases = map[string]string{
	"source":            "Source",
	"name":              "Name",
	"attr":              "Attribute",
	"attribute":         "Attribute",
	"desc":              "Description",
	"description":       "Description",
	"prog":              "Programs",
	"program":           "Programs",
	"programs":          "Programs",
	"mainprogram":       "MainProgram",
	"main_program":      "MainProgram",
	"plat":              "Platforms",
	"platform":          "Platforms",
	"platforms":         "Platforms",
	"pkgset":            "PackageSet",
	"packageset":        "PackageSet",
	"package_set":       "PackageSet",
	"optset":            "OptionSet",
	"optionset":         "OptionSet",
	"option_set":        "OptionSet",
	"ver":               "Version",
	"version":           "Version",
	"loc":               "Loc",
	"parent":            "Parents",
	"parents":           "Parents",
	"related":           "RelatedPackages",
	"relatedpackages":   "RelatedPackages",
	"homepage":          "Homepages",
	"homepages":         "Homepages",
	"id":                "_id",
	"_id":               "_id",
}

func canonicalField(raw string) (string, bool) {
	if mapped, ok := fieldAliases[strings.ToLower(raw)]; ok {
		return mapped, true
	}

	return "", false
}

func tokenizeQuery(input string) []queryToken {
	var tokens []queryToken
	runes := []rune(strings.TrimSpace(input))
	n := len(runes)
	i := 0

	for i < n {
		if unicode.IsSpace(runes[i]) {
			i++

			continue
		}

		switch runes[i] {
		case '(':
			tokens = append(tokens, queryToken{kind: tokLParen, text: "("})
			i++

			continue
		case ')':
			tokens = append(tokens, queryToken{kind: tokRParen, text: ")"})
			i++

			continue
		case '|':
			if i+1 < n && runes[i+1] == '|' {
				i += 2
			} else {
				i++
			}
			tokens = append(tokens, queryToken{kind: tokOr, text: "OR"})

			continue
		case '&':
			if i+1 < n && runes[i+1] == '&' {
				i += 2
			} else {
				i++
			}
			tokens = append(tokens, queryToken{kind: tokAnd, text: "AND"})

			continue
		case '!':
			tokens = append(tokens, queryToken{kind: tokNot, text: "NOT"})
			i++

			continue
		case '+':
			if i+1 < n && !unicode.IsSpace(runes[i+1]) {
				tokens = append(tokens, queryToken{kind: tokPlus, text: "+"})
				i++

				continue
			}
		case '-':
			if i+1 < n && !unicode.IsSpace(runes[i+1]) {
				tokens = append(tokens, queryToken{kind: tokNot, text: "-"})
				i++

				continue
			}
		case '"':
			i++
			start := i
			var sb strings.Builder
			for i < n && runes[i] != '"' {
				if runes[i] == '\\' && i+1 < n {
					i++
				}
				sb.WriteRune(runes[i])
				i++
			}
			if i < n && runes[i] == '"' {
				i++
			}
			phrase := strings.TrimSpace(sb.String())
			if phrase != "" || i > start {
				tokens = append(tokens, queryToken{kind: tokPhrase, text: phrase, quoted: true})
			}

			continue
		}

		// Read a word or field:value token
		var sb strings.Builder
		for i < n && !unicode.IsSpace(runes[i]) && runes[i] != '(' && runes[i] != ')' {
			if runes[i] == '"' {
				// Quoted value after field: e.g. description:"systemd service"
				sb.WriteRune(runes[i])
				i++
				for i < n && runes[i] != '"' {
					if runes[i] == '\\' && i+1 < n {
						sb.WriteRune(runes[i])
						i++
					}
					sb.WriteRune(runes[i])
					i++
				}
				if i < n && runes[i] == '"' {
					sb.WriteRune(runes[i])
					i++
				}

				continue
			}
			sb.WriteRune(runes[i])
			i++
		}

		word := sb.String()
		if word == "" {
			continue
		}

		switch word {
		case "OR":
			tokens = append(tokens, queryToken{kind: tokOr, text: word})

			continue
		case "AND":
			tokens = append(tokens, queryToken{kind: tokAnd, text: word})

			continue
		case "NOT":
			tokens = append(tokens, queryToken{kind: tokNot, text: word})

			continue
		}

		if colonIdx := strings.IndexByte(word, ':'); colonIdx > 0 && colonIdx < len(word)-1 {
			prefix := word[:colonIdx]
			val := word[colonIdx+1:]
			if field, ok := canonicalField(prefix); ok {
				quoted := false
				if len(val) >= 2 && strings.HasPrefix(val, "\"") && strings.HasSuffix(val, "\"") {
					val = strings.TrimSpace(val[1 : len(val)-1])
					quoted = true
				}
				if val != "" {
					tokens = append(tokens, queryToken{
						kind:   tokField,
						field:  field,
						text:   val,
						quoted: quoted,
					})

					continue
				}
			}
		}

		tokens = append(tokens, queryToken{kind: tokWord, text: word})
	}

	return tokens
}

// ParseKeywordQuery parses a user search query into a Bleve Query.
// It supports:
// - Single-word or hyphenated queries (e.g. "git", "home-manager")
// - Multi-word implicit AND queries (e.g. "systemd service")
// - Parenthesized boolean groups (e.g. "(home-manager OR nixpkgs) systemd service")
// - Explicit boolean operators: OR (|), AND (&), NOT (-, !)
// - Case-insensitive field prefixes (e.g. "source:nixos", "name:git", "version:2.47")
// - Quoted phrase queries (e.g. "\"systemd service\"")
// - Raw Bleve query-string modifiers (~, ^, *, ?)
func ParseKeywordQuery(keyword string) (query.Query, []string) {
	trimmed := strings.TrimSpace(keyword)
	if trimmed == "" {
		return bleve.NewMatchAllQuery(), nil
	}

	tokens := tokenizeQuery(trimmed)
	if len(tokens) == 0 {
		return bleve.NewMatchAllQuery(), nil
	}

	// Single plain word without boolean/field operators: preserve exact original Searchix query
	if len(tokens) == 1 && tokens[0].kind == tokWord {
		word := tokens[0].text
		if strings.ContainsAny(word, "~^*?><=/\\") {
			return bleve.NewQueryStringQuery(word), []string{word}
		}

		return buildSingleWordQuery(word), []string{word}
	}

	p := &queryParser{
		tokens: tokens,
		pos:    0,
	}
	q := p.parseOr()
	if q == nil {
		return buildSingleWordQuery(trimmed), []string{trimmed}
	}

	return q, p.positiveTerms
}

func buildSingleWordQuery(word string) query.Query {
	return bleve.NewDisjunctionQuery(
		setBoost(bleve.NewTermQuery(word), 50),
		setBoost(bleve.NewPrefixQuery(word), 25),
		setBoost(bleve.NewMatchPhraseQuery(word), 25),
		bleve.NewMatchQuery(word),
		setBoost(setField(bleve.NewTermQuery(word), "Source"), 40),
	)
}

func buildFieldQuery(field, val string, quoted bool) query.Query {
	if quoted {
		return setField(bleve.NewMatchPhraseQuery(val), field)
	}

	if strings.ContainsAny(val, "~^*?><=/\\") {
		return bleve.NewQueryStringQuery(field + ":" + val)
	}

	return bleve.NewDisjunctionQuery(
		setBoost(setField(bleve.NewTermQuery(val), field), 25),
		setBoost(setField(bleve.NewPrefixQuery(val), field), 10),
		setField(bleve.NewMatchPhraseQuery(val), field),
		setField(bleve.NewMatchQuery(val), field),
	)
}

type queryParser struct {
	tokens        []queryToken
	pos           int
	positiveTerms []string
}

func (p *queryParser) peek() queryToken {
	if p.pos >= len(p.tokens) {
		return queryToken{kind: tokEOF}
	}

	return p.tokens[p.pos]
}

func (p *queryParser) next() queryToken {
	tok := p.peek()
	if p.pos < len(p.tokens) {
		p.pos++
	}

	return tok
}

func (p *queryParser) parseOr() query.Query {
	left := p.parseAnd()
	if left == nil {
		return nil
	}

	disjuncts := []query.Query{left}
	for p.peek().kind == tokOr {
		p.next() // consume OR
		right := p.parseAnd()
		if right != nil {
			disjuncts = append(disjuncts, right)
		}
	}

	if len(disjuncts) == 1 {
		return disjuncts[0]
	}

	return bleve.NewDisjunctionQuery(disjuncts...)
}

type unaryClause struct {
	q       query.Query
	negated bool
}

func (p *queryParser) parseAnd() query.Query {
	var clauses []unaryClause

	for {
		tok := p.peek()
		if tok.kind == tokEOF || tok.kind == tokRParen || tok.kind == tokOr {
			break
		}
		if tok.kind == tokAnd {
			p.next() // consume explicit AND

			continue
		}

		clause, ok := p.parseUnary()
		if !ok || clause.q == nil {
			break
		}
		clauses = append(clauses, clause)
	}

	if len(clauses) == 0 {
		return nil
	}

	if len(clauses) == 1 && !clauses[0].negated {
		return clauses[0].q
	}

	b := bleve.NewBooleanQuery()
	positiveCount := 0
	for _, c := range clauses {
		if c.negated {
			b.AddMustNot(c.q)
		} else {
			b.AddMust(c.q)
			positiveCount++
		}
	}

	if positiveCount == 0 {
		b.AddMust(bleve.NewMatchAllQuery())
	}

	return b
}

func (p *queryParser) parseUnary() (unaryClause, bool) {
	negated := false
	for {
		tok := p.peek()
		if tok.kind == tokNot {
			p.next()
			negated = !negated

			continue
		}
		if tok.kind == tokPlus {
			p.next()

			continue
		}

		break
	}

	tok := p.peek()
	switch tok.kind {
	case tokEOF, tokRParen, tokOr, tokAnd:
		return unaryClause{}, false
	case tokLParen:
		p.next() // consume '('
		inner := p.parseOr()
		if p.peek().kind == tokRParen {
			p.next() // consume ')'
		}
		if inner == nil {
			return p.parseUnary()
		}

		return unaryClause{q: inner, negated: negated}, true
	case tokPhrase:
		p.next()
		if !negated && tok.text != "" {
			p.positiveTerms = append(p.positiveTerms, tok.text)
		}

		return unaryClause{
			q: bleve.NewDisjunctionQuery(
				setBoost(bleve.NewMatchPhraseQuery(tok.text), 25),
				bleve.NewMatchQuery(tok.text),
			),
			negated: negated,
		}, true
	case tokField:
		p.next()
		if !negated && tok.text != "" {
			p.positiveTerms = append(p.positiveTerms, tok.text)
		}

		return unaryClause{
			q:       buildFieldQuery(tok.field, tok.text, tok.quoted),
			negated: negated,
		}, true
	case tokWord:
		p.next()
		word := tok.text
		if !negated && word != "" {
			p.positiveTerms = append(p.positiveTerms, word)
		}
		if strings.ContainsAny(word, "~^*?><=/\\") {
			return unaryClause{
				q:       bleve.NewQueryStringQuery(word),
				negated: negated,
			}, true
		}

		return unaryClause{
			q:       buildSingleWordQuery(word),
			negated: negated,
		}, true
	default:
		p.next()

		return unaryClause{}, false
	}
}
