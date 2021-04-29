package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"gopkg.in/yaml.v3"
	"k8s.io/test-infra/prow/config/org"
)

type Sigs struct {
	Sigs       []*Group `yaml:"sigs"`
	Usergroups []*Group `yaml:"usergroups"`
	Committees []*Group `yaml:"committees"`
}

type Group struct {
	Dir              string
	Name             string
	MissionStatement string         `yaml:"mission_statement,omitempty"`
	Label            string         `yaml:",omitempty"`
	Leadership       *Leadership    `yaml:",omitempty"`
	Meetings         []*Meeting     `yaml:",omitempty"`
	Contact          *Contact       `yaml:",omitempty"`
	SubProjects      []*SubProjects `yaml:",omitempty"`
}

type Contact struct {
	Slack       string  `yaml:"slack"`
	MailingList string  `yaml:"mailing_list"`
	Teams       []*Team `yaml:"teams"`
}

type Team struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

type Meeting struct {
	Description   string
	Day           string
	Time          string
	TZ            string
	Frequency     string
	URL           string
	ArchiveURL    string `yaml:"archive_url"`
	RecordingsURL string `yaml:"recordings_url"`
}

type Leadership struct {
	Chairs []*Chair
}

type Chair struct {
	Github  string
	Name    string
	Company string
}

type SubProjects struct {
	Name   string
	Owners []string
}

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
	fs.StringVar(&o.sigsFilePath, "sigs_file_path", "", "File path to the sigs.yaml file to check.")
	err := fs.Parse(os.Args[1:])
	if err != nil {
		log.Fatalf("error parsing arguments %v: %v", os.Args[1:], err)
	}
	return o
}

const DefaultURIForKubevirtOrgsYAML = "https://raw.githubusercontent.com/kubevirt/project-infra/master/github/ci/prow-deploy/files/orgs.yaml"

func main() {
	options := gatherOptions()
	if err := options.Validate(); err != nil {
		log.Fatalf("invalid arguments: %v", err)
	}
	log.Printf("dry-run: %v", options.dryRun)

	orgs, err := getOrgs()
	org := "kubevirt"
	generatedBaseOwnersFileURIs := map[string]struct{}{
		"https://raw.githubusercontent.com/k8snetworkplumbingwg/kubemacpool/master/OWNERS": {},
		"https://raw.githubusercontent.com/k8snetworkplumbingwg/multi-networkpolicy-iptables/master/OWNERS": {},
		"https://raw.githubusercontent.com/k8snetworkplumbingwg/sriov-network-operator/master/OWNERS": {},
		"https://raw.githubusercontent.com/nmstate/kubernetes-nmstate/master/OWNERS": {},
		"https://raw.githubusercontent.com/virtblocks/kubevirt/master/OWNERS": {},
	}
	missingBaseOwnersFileURIs := make(map[string]struct{}, 0)
	for repoName, _ := range orgs.Orgs[org].Repos {
		generatedBaseOwnersFileURIs[fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/master/OWNERS", org, repoName)] = struct{}{}
	}

	buf, err := ioutil.ReadFile(options.sigsFilePath)
	if err != nil {
		log.Fatalf("error reading %s: %v", options.sigsFilePath, err)
	}

	sigs := &Sigs{}
	err = yaml.Unmarshal(buf, sigs)
	if err != nil {
		log.Fatalf("in file %q: %v", options.sigsFilePath, err)
	}

	for _, sig := range sigs.Sigs {
		for _, subProject := range sig.SubProjects {
			log.Printf("checking sig %s subproject %s", sig.Name, subProject.Name)
			foundOwners := make([]string, 0)
			for _, ownersFileURL := range subProject.Owners {
				response, err := http.DefaultClient.Head(ownersFileURL)
				if _, exists := generatedBaseOwnersFileURIs[ownersFileURL]; !exists {
					missingBaseOwnersFileURIs[ownersFileURL] = struct{}{}
				}
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

	if len(missingBaseOwnersFileURIs) > 0 {
		log.Fatalf("missing links to repos in file %q: %v", options.sigsFilePath, missingBaseOwnersFileURIs)
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

func getOrgs() (*org.FullConfig, error) {
	response, err := http.DefaultClient.Get(DefaultURIForKubevirtOrgsYAML)
	if err != nil {
		return nil, fmt.Errorf("could not get %q: %v", DefaultURIForKubevirtOrgsYAML, err)
	}
	if response.StatusCode >= 300 || response.StatusCode < http.StatusOK {
		return nil, fmt.Errorf("could not get %q: %v", DefaultURIForKubevirtOrgsYAML, response.Status)
	}
	defer response.Body.Close()

	var cfg org.FullConfig
	if err := yaml.NewDecoder(response.Body).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("failed to load %q: %v", DefaultURIForKubevirtOrgsYAML, err)
	}
	return &cfg, err
}
