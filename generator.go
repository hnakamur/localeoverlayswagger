package multiswagger

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/ghodss/yaml"

	"github.com/goadesign/goa/design"
	"github.com/goadesign/goa/goagen/codegen"
	"github.com/goadesign/goa/goagen/gen_swagger"
	"github.com/goadesign/goa/goagen/utils"
)

// Generator is the swagger code generator.
type Generator struct {
	API        *design.APIDefinition // The API definition
	OutDir     string                // Path to output directory
	LocalesDir string                // Path to locales directory
	genfiles   []string              // Generated files
}

// Generate is the generator entry point called by the meta generator.
func Generate() (files []string, err error) {
	var outDir, localesDir, ver string
	set := flag.NewFlagSet("swagger", flag.PanicOnError)
	set.StringVar(&outDir, "out", "", "")
	set.StringVar(&localesDir, "locales", "locales", "")
	set.StringVar(&ver, "version", "", "")
	set.String("design", "", "")
	set.Parse(os.Args[1:])

	if err := codegen.CheckVersion(ver); err != nil {
		return nil, err
	}

	g := &Generator{OutDir: outDir, LocalesDir: localesDir, API: design.Design}

	return g.Generate()
}

// Generate produces the skeleton main.
func (g *Generator) Generate() (_ []string, err error) {
	go utils.Catch(nil, func() { g.Cleanup() })

	defer func() {
		if err != nil {
			g.Cleanup()
		}
	}()

	swagger, err := genswagger.New(g.API)
	if err != nil {
		return nil, err
	}

	swaggerDir := filepath.Join(g.OutDir, "swagger")
	os.RemoveAll(swaggerDir)
	if err = os.MkdirAll(swaggerDir, 0755); err != nil {
		return nil, err
	}
	g.genfiles = append(g.genfiles, swaggerDir)

	// JSON
	rawJSON, err := json.Marshal(swagger)
	if err != nil {
		return nil, err
	}
	swaggerFile := filepath.Join(swaggerDir, "swagger.json")
	if err := ioutil.WriteFile(swaggerFile, rawJSON, 0644); err != nil {
		return nil, err
	}
	g.genfiles = append(g.genfiles, swaggerFile)

	// YAML
	rawYAML, err := yaml.JSONToYAML(rawJSON)
	if err != nil {
		return nil, err
	}
	swaggerFile = filepath.Join(swaggerDir, "swagger.yaml")
	if err := ioutil.WriteFile(swaggerFile, rawYAML, 0644); err != nil {
		return nil, err
	}
	g.genfiles = append(g.genfiles, swaggerFile)

	localeFilePaths, err := filepath.Glob(filepath.Join(g.LocalesDir, "*.yaml"))
	for _, localeFilePath := range localeFilePaths {
		localeFilename := filepath.Base(localeFilePath)
		locale := localeFilename[:len(localeFilename)-len(".yaml")]

		rawLocaleYAML, err := ioutil.ReadFile(localeFilePath)
		if err != nil {
			return nil, err
		}
		rawLocaleJSON, err := yaml.YAMLToJSON(rawLocaleYAML)
		if err != nil {
			return nil, err
		}
		var localeJSON map[string]interface{}
		if err = json.Unmarshal(rawLocaleJSON, &localeJSON); err != nil {
			return nil, err
		}

		var mergedJSON map[string]interface{}
		if err := decodeJSONUsingNumber(rawJSON, &mergedJSON); err != nil {
			return nil, err
		}
		mergeMapsRecursive(mergedJSON, localeJSON)

		// JSON
		mergedRawJSON, err := json.Marshal(mergedJSON)
		if err != nil {
			return nil, err
		}
		swaggerFile = filepath.Join(swaggerDir, fmt.Sprintf("swagger.%s.json", locale))
		if err := ioutil.WriteFile(swaggerFile, mergedRawJSON, 0644); err != nil {
			return nil, err
		}
		g.genfiles = append(g.genfiles, swaggerFile)

		// YAML
		mergedRawYAML, err := yaml.JSONToYAML(mergedRawJSON)
		if err != nil {
			return nil, err
		}
		swaggerFile := filepath.Join(swaggerDir, fmt.Sprintf("swagger.%s.yaml", locale))
		if err := ioutil.WriteFile(swaggerFile, mergedRawYAML, 0644); err != nil {
			return nil, err
		}
		g.genfiles = append(g.genfiles, swaggerFile)
	}
	return g.genfiles, nil
}

// Cleanup removes all the files generated by this generator during the last invokation of Generate.
func (g *Generator) Cleanup() {
	for _, f := range g.genfiles {
		os.Remove(f)
	}
	g.genfiles = nil
}

// decodeJSONUsingNumber decodes a JSON using Number instead of float64 for a number.
func decodeJSONUsingNumber(data []byte, v interface{}) error {
	decoder := json.NewDecoder(bytes.NewBuffer(data))
	decoder.UseNumber()
	return decoder.Decode(v)
}

func mergeMapsRecursive(dest, src map[string]interface{}) {
	for k, v := range src {
		srcMap, srcIsMap := v.(map[string]interface{})
		if srcIsMap {
			destMap, destIsMap := dest[k].(map[string]interface{})
			if destIsMap {
				mergeMapsRecursive(destMap, srcMap)
				continue
			}
		}
		dest[k] = v
	}
}