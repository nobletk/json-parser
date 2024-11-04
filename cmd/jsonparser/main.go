package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/nobletk/json-parser/internal/lexer"
	"github.com/nobletk/json-parser/internal/parser"
	"github.com/nobletk/json-parser/pkg/mylog"
	"github.com/spf13/pflag"
)

type config struct {
	debug bool
}

func main() {
	var cfg config

	pflag.BoolVarP(&cfg.debug, "debug", "d", false, "debug mode for logs")
	pflag.Usage = func() {
		var buf bytes.Buffer

		buf.WriteString("Usage:\n")
		buf.WriteString(" jsonparser [OPTIONS] <FILEPATH>\n")
		buf.WriteString(" cat <FILEPATH> | jsonparser [OPTIONS]\n")

		fmt.Fprintf(os.Stderr, buf.String())
		pflag.PrintDefaults()
	}
	pflag.Parse()

	if len(pflag.Args()) > 1 {
		pflag.Usage()
		os.Exit(2)
	}

	logger := mylog.CreateLogger(cfg.debug)

	filePath := pflag.Arg(0)
	data, err := readData(filePath)
	if err != nil {
		log.Fatal(err)
	}

	var out bytes.Buffer

	out.WriteString("Data:\n")
	out.WriteString(fmt.Sprintf("%s\n\n", string(data)))

	l := lexer.New(logger, string(data))
	p := parser.New(l)
	parsedJSON, jsonErr := p.ParseFile()
	if jsonErr != nil {
		out.WriteString("Invalid JSON:\n")
		out.WriteString(fmt.Sprintf("    %s", jsonErr.Msg))
		out.WriteString(fmt.Sprintf("    Position(line %d, column %d)\n", jsonErr.Pos.Line,
			jsonErr.Pos.Column))

		fmt.Print(out.String())
		os.Exit(1)
	}

	validJSON, err := json.MarshalIndent(parsedJSON.ToInterface(), "", "  ")
	if err != nil {
		fmt.Printf("MarshalIndent() Failed. %s\n", err)
		os.Exit(1)
	}

	out.WriteString("Valid JSON:\n")
	out.WriteString(fmt.Sprintf("%s\n", string(validJSON)))

	fmt.Print(out.String())
	os.Exit(0)
}

func readData(filePath string) ([]byte, error) {
	if filePath == "" {
		r, err := io.ReadAll(os.Stdin)
		if err != nil {
			return nil, err
		}
		return r, nil
	} else {
		r, err := os.ReadFile(filePath)
		if err != nil {
			return nil, err
		}
		return r, nil
	}
}
