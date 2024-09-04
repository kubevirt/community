package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"kubevirt.io/community/pkg/sigs"
	"log"
	"net/http"
	"os"

	yaml "gopkg.in/yaml.v3"
)

type options struct {
	dryRun       bool
	sigsFilePath string
}

func (o *options) Validate() error {
	if o.sigsFilePath == "" {
		return fmt.Errorf("path to sigs.yaml is required")
	}
	if _, err := os.Stat(o.sigsFilePath); os.IsNotExist(err) {
		return fmt.Errorf("file %s does not exist", o.sigsFilePath)
	}
	return nil
}

func gatherOptions() options {
	o := options{}
	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	fs.BoolVar(&o.dryRun, "dry-run", true, "Dry run for testing. Uses API tokens but does not mutate.")
	fs.StringVar(&o.sigsFilePath, "sigs_file_path", "./sigs.yaml", "File path to the sigs.yaml file to check.")
	err := fs.Parse(os.Args[1:])
	if err != nil {
		log.Fatalf("error parsing arguments %v: %v", os.Args[1:], err)
	}
	return o
}

func main() {
	options := gatherOptions()
	if err := options.Validate(); err != nil {
		log.Fatalf("invalid arguments: %v", err)
	}
	log.Printf("dry-run: %v", options.dryRun)

	sigs, err := sigs.ReadFile(options.sigsFilePath)
	if err != nil {
		log.Fatalf("invalid arguments: %v", err)
	}

	for _, sig := range sigs.Sigs {
		for _, subProject := range sig.SubProjects {
			log.Printf("checking sig %s subproject %s", sig.Name, subProject.Name)
			foundOwners := make([]string, 0)
			for _, ownersFileURL := range subProject.Owners {
				response, err := http.DefaultClient.Head(ownersFileURL)
				if err != nil {
					log.Printf("failed to retrieve %q, continuing with next", ownersFileURL)
					continue
				}
				defer response.Body.Close()
				if response.StatusCode >= http.StatusOK && response.StatusCode < http.StatusMultipleChoices {
					foundOwners = append(foundOwners, ownersFileURL)
				} else {
					log.Printf("failed to retrieve %q: %d", ownersFileURL, response.StatusCode)
				}
			}
			subProject.Owners = foundOwners
		}
	}

	output, err := yaml.Marshal(sigs)
	if err != nil {
		log.Fatalf("in file %q: %v", options.sigsFilePath, err)
	}
	if options.dryRun {
		_, err := os.Stdout.Write(output)
		if err != nil {
			log.Fatalf("file %q: %v", options.sigsFilePath, err)
		}
	} else {
		stat, err := os.Stat(options.sigsFilePath)
		if err != nil {
			log.Fatalf("stat for file %q failed: %v", options.sigsFilePath, err)
		}
		err = ioutil.WriteFile(options.sigsFilePath, output, stat.Mode())
		if err != nil {
			log.Fatalf("write to file %q failed: %v", options.sigsFilePath, err)
		}
	}

}
