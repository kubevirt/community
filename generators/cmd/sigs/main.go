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
	"github.com/sirupsen/logrus"
	"kubevirt.io/community/pkg/sigs"
	"os"
	"text/template"
)

var (
	debug        bool
	sigsFilePath string
	//go:embed sig-list.gomd
	sigListTemplate string
)

type SigListTemplateData struct {
	Sigs *sigs.Sigs
}

func init() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
}

func main() {
	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	fs.BoolVar(&debug, "debug", false, "whether debug information should be printed")
	fs.StringVar(&sigsFilePath, "sigs-file-path", "./sigs.yaml", "file path to the sigs.yaml file to check")
	err := fs.Parse(os.Args[1:])
	if err != nil {
		log().Fatalf("error parsing arguments %v: %v", os.Args[1:], err)
	}

	if debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	sigs, err := sigs.ReadFile(sigsFilePath)
	if err != nil {
		log().Fatalf("invalid arguments: %v", err)
	}

	sigListFile, err := os.Create("./sig-list.md")
	if err != nil {
		log().Fatalf("could not open file: %v", err)
	}
	defer sigListFile.Close()

	t, err := template.New("report").Parse(sigListTemplate)
	if err != nil {
		log().Fatalf("failed to load template: %v", err)
	}

	err = t.Execute(sigListFile, SigListTemplateData{Sigs: sigs})
	if err != nil {
		log().Fatalf("failed to execute template: %v", err)
	}
}

func log() *logrus.Entry {
	return logrus.WithFields(
		logrus.Fields{
			"generator": "sigs",
		},
	)
}
