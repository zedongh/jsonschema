package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/santhosh-tekuri/jsonschema/v5"
	_ "github.com/santhosh-tekuri/jsonschema/v5/httploader"
	"gopkg.in/yaml.v2"
)

func usage() {
	fmt.Fprintln(os.Stderr, "jv [-draft INT] [-output FORMAT] [-assertformat] [-assertcontent] <json-schema> [<json-or-yaml-doc>]...")
	flag.PrintDefaults()
}

func main() {
	draft := flag.Int("draft", 2020, "draft used when '$schema' attribute is missing. valid values 4, 5, 7, 2019, 2020")
	output := flag.String("output", "", "output format. valid values flag, basic, detailed")
	assertFormat := flag.Bool("assertformat", false, "enable format assertions with draft >= 2019")
	assertContent := flag.Bool("assertcontent", false, "enable content assertions with draft >= 2019")
	flag.Usage = usage
	flag.Parse()
	if len(flag.Args()) == 0 {
		usage()
		os.Exit(1)
	}

	compiler := jsonschema.NewCompiler()
	switch *draft {
	case 4:
		compiler.Draft = jsonschema.Draft4
	case 6:
		compiler.Draft = jsonschema.Draft6
	case 7:
		compiler.Draft = jsonschema.Draft7
	case 2019:
		compiler.Draft = jsonschema.Draft2019
	case 2020:
		compiler.Draft = jsonschema.Draft2020
	default:
		fmt.Fprintln(os.Stderr, "draft must be 4, 5, 7, 2019 or 2020")
		os.Exit(1)
	}

	compiler.AssertFormat = *assertFormat
	compiler.AssertContent = *assertContent

	var validOutput bool
	for _, out := range []string{"", "flag", "basic", "detailed"} {
		if *output == out {
			validOutput = true
			break
		}
	}
	if !validOutput {
		fmt.Fprintln(os.Stderr, "output must be flag, basic or detailed")
		os.Exit(1)
	}

	schema, err := compiler.Compile(flag.Arg(0))
	if err != nil {
		fmt.Fprintf(os.Stderr, "%#v\n", err)
		os.Exit(1)
	}

	for _, f := range flag.Args()[1:] {
		file, err := os.Open(f)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
		defer file.Close()

		v, err := decodeFile(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}

		err = schema.Validate(v)
		if err != nil {
			if ve, ok := err.(*jsonschema.ValidationError); ok {
				var out interface{}
				switch *output {
				case "flag":
					out = ve.FlagOutput()
				case "basic":
					out = ve.BasicOutput()
				case "detailed":
					out = ve.DetailedOutput()
				}
				if out == nil {
					fmt.Fprintf(os.Stderr, "%#v\n", err)
				} else {
					b, _ := json.MarshalIndent(out, "", "  ")
					fmt.Fprintln(os.Stderr, string(b))
				}
			} else {
				fmt.Fprintf(os.Stderr, "validation failed: %v\n", err)
			}
			os.Exit(1)
		}
	}
}

func decodeFile(file *os.File) (interface{}, error) {
	ext := filepath.Ext(file.Name())
	if ext == ".yaml" || ext == ".yml" {
		var v interface{}
		dec := yaml.NewDecoder(file)
		if err := dec.Decode(&v); err != nil {
			return nil, fmt.Errorf("invalid yaml file %s: %v", file.Name(), err)
		}
		v, err := toStringKeys(v)
		if err != nil {
			return nil, fmt.Errorf("error converting %s to json: %v", file.Name(), err)
		}
		return v, nil
	}

	// json file
	var v interface{}
	dec := json.NewDecoder(file)
	dec.UseNumber()
	if err := dec.Decode(&v); err != nil {
		return nil, fmt.Errorf("invalid json file %s: %v", file.Name(), err)
	}
	return v, nil
}

func toStringKeys(val interface{}) (interface{}, error) {
	var err error
	switch val := val.(type) {
	case map[interface{}]interface{}:
		m := make(map[string]interface{})
		for key, v := range val {
			k, ok := key.(string)
			if !ok {
				return nil, fmt.Errorf("found non-string key: %v", key)
			}
			m[k], err = toStringKeys(v)
			if err != nil {
				return nil, err
			}
		}
		return m, nil
	case []interface{}:
		var l = make([]interface{}, len(val))
		for i, v := range val {
			l[i], err = toStringKeys(v)
			if err != nil {
				return nil, err
			}
		}
		return l, nil
	default:
		return val, nil
	}
}
