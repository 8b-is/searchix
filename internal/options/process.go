package options

import (
	"encoding/json"
	"io"
	"log/slog"
	"os"

	"github.com/bcicen/jstream"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
)

type nixValueJSON struct {
	Type string `mapstructure:"_type"`
	Text string
}

type linkJSON struct {
	Name string
	URL  string `json:"url"`
}

type nixOptionJSON struct {
	Declarations    []linkJSON
	Default         nixValueJSON
	Description     string
	Example         nixValueJSON
	Loc             []string
	ReadOnly        bool
	RelatedPackages string
	Type            string
}

func ValueTypeToString(valueType jstream.ValueType) string {
	switch valueType {
	case jstream.Unknown:
		return "unknown"
	case jstream.Null:
		return "null"
	case jstream.String:
		return "string"
	case jstream.Number:
		return "number"
	case jstream.Boolean:
		return "boolean"
	case jstream.Array:
		return "array"
	case jstream.Object:
		return "object"
	}

	return "very strange"
}

func convertNixValue(nj nixValueJSON) *NixValue {
	switch nj.Type {
	case "", "literalExpression":
		if nj.Text == "" {
			return nil
		}

		return &NixValue{
			Text: nj.Text,
		}
	case "literalMD":
		return &NixValue{
			Markdown: Markdown(nj.Text),
		}
	default:
		slog.Warn("got unexpected NixValue type", "type", nj.Type, "text", nj.Text)

		return nil
	}
}

func Process(inpath string, outpath string) error {
	infile, err := os.Open(inpath)
	if err != nil {
		return errors.WithMessagef(err, "failed to open input file %s", inpath)
	}
	defer infile.Close()
	outfile, err := os.Create(outpath)
	if err != nil {
		return errors.WithMessagef(err, "failed to open output file %s", outpath)
	}
	if outpath != "/dev/stdout" {
		defer outfile.Close()
	}

	dec := jstream.NewDecoder(infile, 1).EmitKV()
	var optJSON nixOptionJSON
	ms, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		ErrorUnused: true,
		ZeroFields:  true,
		Result:      &optJSON,
		Squash:      true,
		DecodeHook:  mapstructure.TextUnmarshallerHookFunc(),
	})
	if err != nil {
		return errors.WithMessage(err, "could not create mapstructure decoder")
	}

	_, err = outfile.WriteString("[\n")
	if err != nil {
		return errors.WithMessage(err, "could not write to output")
	}
	for mv := range dec.Stream() {
		if err := dec.Err(); err != nil {
			return errors.WithMessage(err, "could not decode JSON")
		}
		if mv.ValueType != jstream.Object {
			return errors.Errorf("unexpected object type %s", ValueTypeToString(mv.ValueType))
		}
		kv := mv.Value.(jstream.KV)
		x := kv.Value.(map[string]interface{})

		err = ms.Decode(x) // stores in optJSON
		if err != nil {
			return errors.WithMessagef(err, "failed to decode option %#v", x)
		}

		var decs = make([]Link, len(optJSON.Declarations))
		for i, d := range optJSON.Declarations {
			decs[i] = Link(d)
		}

		opt := NixOption{
			Option:          kv.Key,
			Declarations:    decs,
			Default:         convertNixValue(optJSON.Default),
			Description:     Markdown(optJSON.Description),
			Example:         convertNixValue(optJSON.Example),
			RelatedPackages: Markdown(optJSON.RelatedPackages),
			Loc:             optJSON.Loc,
			Type:            optJSON.Type,
		}

		b, err := json.MarshalIndent(opt, "", "  ")
		if err != nil {
			return errors.WithMessagef(err, "failed to encode option %#v", opt)
		}

		_, err = outfile.Write(b)
		if err != nil {
			return errors.WithMessage(err, "failed to write to output")
		}
		_, err = outfile.WriteString(",\n")
		if err != nil {
			return errors.WithMessage(err, "failed to write to output")
		}
	}

	if outpath != "/dev/stdout" {
		_, err = outfile.Seek(-2, io.SeekCurrent)
		if err != nil {
			return errors.WithMessage(err, "could not write to output")
		}
	}

	_, err = outfile.WriteString("]\n")
	if err != nil {
		return errors.WithMessage(err, "could not write to output")
	}

	return nil
}
