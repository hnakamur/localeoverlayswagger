# LocaleOverlaySwagger

A [goa](https://github.com/goadesign/goa) plugin package to generate localized swagger specification files.

## Overview

Multiswagger is compatible with internal swagger generator of goagen (`goagen swagger`) but it expects additional locale YAML files to overlay fields like description.

The path of locale YAML files are `locales/*.yaml`.
For example, the path for the Japanese locale YAML is `locales/ja.yaml`.
The locale YAML file contains only fields to overlay.

Multiswagger generates `swagger.${locale}.json` and `swagger.${locale}.yaml` as well as the original `swagger.json` and `swagger.yaml`.

See https://github.com/hnakamur/goa-getting-started/tree/overlay_japanese_yaml for an example.

## Installation

```sh
$ go get github.com/hnakmaur/localeoverlayswagger
```

## Usage


```sh
$ goagen gen --pkg-path github.com/hnakmaur/localeoverlayswagger --design package/path/to/your/design
```

Please add `--locales locales_dir` if your `locales_dir` is different from the default value `locales`.

## License

MIT License
