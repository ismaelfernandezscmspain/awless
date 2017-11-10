/*
Copyright 2017 WALLIX

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"bytes"
	"io/ioutil"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/wallix/awless/gen/aws"
)

func generateDriverTypes() {
	templ, err := template.New("types").Funcs(template.FuncMap{
		"Title":          strings.Title,
		"ToUpper":        strings.ToUpper,
		"ApiToInterface": aws.ApiToInterface,
	}).Parse(typesTempl)
	if err != nil {
		panic(err)
	}

	var buff bytes.Buffer
	err = templ.Execute(&buff, aws.DriversDefs)
	if err != nil {
		panic(err)
	}

	if err := ioutil.WriteFile(filepath.Join(DRIVERS_DIR, "gen_drivers.go"), buff.Bytes(), 0666); err != nil {
		panic(err)
	}
}

const typesTempl = `/* Copyright 2017 WALLIX

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// DO NOT EDIT
// This file was automatically generated with go generate
package awsdriver

import (
	"strings"
	"github.com/wallix/awless/template/driver"
	"github.com/wallix/awless/logger"
	{{- range $index, $service := . }}
  "github.com/aws/aws-sdk-go/service/{{ $service.Api }}/{{ $service.Api }}iface"
	{{- end }}
)

{{ range $, $service := . }}
type {{ Title $service.Api }}Driver struct {
	dryRun bool
	logger *logger.Logger
	{{ $service.Api }}iface.{{ ApiToInterface $service.Api }}
}

func (d *{{ Title $service.Api }}Driver) SetDryRun(dry bool)         { d.dryRun = dry }
func (d *{{ Title $service.Api }}Driver) SetLogger(l *logger.Logger) { d.logger = l }
func New{{ Title $service.Api }}Driver(api {{ $service.Api }}iface.{{ ApiToInterface $service.Api }}) driver.Driver{
	return &{{ Title $service.Api }}Driver{false, logger.DiscardLogger, api}
}

func (d *{{ Title $service.Api }}Driver) LookupIface(lookups ...string) (interface{}, error) {
	return nil, nil
}

func (d *{{ Title $service.Api }}Driver) Lookup(lookups ...string) (driverFn driver.DriverFn, err error) {
	return nil, driver.ErrDriverFnNotFound
}

{{ end }}`
