package options

import (
	"encoding/json"
	"io"
	"os"

	"github.com/bcicen/jstream"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
)

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
	var opt NixOption
	ms, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		ErrorUnused: true,
		Result:      &opt,
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
		if kv.Key == "_module.args" {
			continue
		}
		x := kv.Value.(map[string]interface{})
		x["option"] = kv.Key

		err = ms.Decode(x)
		if err != nil {
			return errors.WithMessagef(err, "failed to decode option %#v", x)
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

	_, err = outfile.Seek(-2, io.SeekCurrent)
	if err != nil {
		return errors.WithMessage(err, "could not write to output")
	}

	_, err = outfile.WriteString("]\n")
	if err != nil {
		return errors.WithMessage(err, "could not write to output")
	}

	return nil
}
