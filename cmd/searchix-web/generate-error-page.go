package main

import (
	"io"
	"os"
	"strings"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fmsg"

	"alin.ovh/searchix/frontend"
	"alin.ovh/searchix/internal/components"
	"alin.ovh/searchix/internal/config"
	"alin.ovh/searchix/internal/file"
)

const perm = 0o644

type GenerateErrorPageOptions struct {
	Outdir string `long:"outdir" description:"output directory"`
}

func (opts *GenerateErrorPageOptions) Execute(_ []string) error {
	assets, err := frontend.New()
	if err != nil {
		return fault.Wrap(err, fmsg.With("could not create frontend"))
	}

	root, err := file.CreateAndOpenRoot(opts.Outdir)
	if err != nil {
		return fault.Wrap(err, fmsg.With("could not open root directory"))
	}
	defer root.Close()

	err = writeErrorPageTemplate(root, assets)
	if err != nil {
		return err
	}

	err = root.Mkdir("assets", 0o755)
	if err != nil {
		return fault.Wrap(err, fmsg.With("could not create assets directory"))
	}

	for _, asset := range assets.ByPath {
		if err := writeFile(root, asset); err != nil {
			return err
		}
	}

	return nil
}

func writeErrorPageTemplate(root *file.Root, assets *frontend.AssetCollection) error {
	out, err := root.OpenFile("error.html", os.O_CREATE|os.O_WRONLY, perm)
	if err != nil {
		return fault.Wrap(err, fmsg.With("could not open error.html"))
	}
	defer out.Close()

	err = components.ErrorTemplate(components.TemplateData{
		Source: config.Source{
			Importer: config.All,
		},
		Sources:       []config.Source{},
		Query:         "",
		ExtraHeadHTML: "",
		Code:          0,
		Message:       `{{placeholder "http.error.status_code"}} {{placeholder "http.error.status_text"}}`,
		Assets:        assets,
	}).Render(out)
	if err != nil {
		return fault.Wrap(err, fmsg.With("could not render error template"))
	}

	err = out.Sync()
	if err != nil {
		return fault.Wrap(err, fmsg.With("could not sync error.html"))
	}

	return nil
}

func writeFile(root *file.Root, asset *frontend.Asset) error {
	outname := strings.TrimPrefix(asset.ImmutablePath, "/")

	in, err := frontend.Files.Open(asset.Filename)
	if err != nil {
		return fault.Wrap(err, fmsg.With("could not open asset file"))
	}
	defer in.Close()

	out, err := root.OpenFile(outname, os.O_CREATE|os.O_WRONLY, perm)
	if err != nil {
		return fault.Wrap(err, fmsg.With("could not open output file"))
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return fault.Wrap(err, fmsg.With("could not write asset file"))
	}

	err = out.Sync()
	if err != nil {
		return fault.Wrap(err, fmsg.With("could not sync asset file"))
	}

	return nil
}

func init() {
	var opts GenerateErrorPageOptions

	_, err := parser.AddCommand(
		"generate-error-page",
		"generate error page template",
		"generate error page template for use with a reverse proxy",
		&opts,
	)
	if err != nil {
		panic(err)
	}
}
