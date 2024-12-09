/*
 * This file is part of the KubeVirt project
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * Copyright the KubeVirt Authors.
 *
 */

package main

import (
	_ "embed"
	"flag"
	"fmt"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"os"
	"text/template"
)

var (
	debug          bool
	alumniFilePath string
	//go:embed ALUMNI.gomd
	alumniTemplate string
)

type Alumni struct {
	Github  string
	Name    string `yaml:",omitempty"`
	Company string `yaml:",omitempty"`
	Since   string `yaml:",omitempty"`
}

type AlumniTemplateData struct {
	Alumni []*Alumni
}

func init() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
}

func main() {
	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	fs.BoolVar(&debug, "debug", false, "whether debug information should be printed")
	fs.StringVar(&alumniFilePath, "alumni-file-path", "./alumni.yaml", "file path to the alumni.yaml file")
	err := fs.Parse(os.Args[1:])
	if err != nil {
		log().Fatalf("error parsing arguments %v: %v", os.Args[1:], err)
	}

	if debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	alumniTemplateData, err := readFile(alumniFilePath)
	if err != nil {
		log().Fatalf("invalid arguments: %v", err)
	}

	alumniFile, err := os.Create("./ALUMNI.md")
	if err != nil {
		log().Fatalf("could not open file: %v", err)
	}
	defer alumniFile.Close()

	t, err := template.New("report").Parse(alumniTemplate)
	if err != nil {
		log().Fatalf("failed to load template: %v", err)
	}

	err = t.Execute(alumniFile, alumniTemplateData)
	if err != nil {
		log().Fatalf("failed to execute template: %v", err)
	}
}

func readFile(path string) (*AlumniTemplateData, error) {
	buf, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading %s: %v", path, err)
	}

	alumni := &AlumniTemplateData{}
	err = yaml.Unmarshal(buf, alumni)
	if err != nil {
		return nil, fmt.Errorf("in file %q: %v", path, err)
	}
	return alumni, err
}

func log() *logrus.Entry {
	return logrus.WithFields(
		logrus.Fields{
			"generator": "sigs",
		},
	)
}
