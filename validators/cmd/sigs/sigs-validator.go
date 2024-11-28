package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"kubevirt.io/community/pkg/labels"
	"kubevirt.io/community/pkg/orgs"
	"kubevirt.io/community/pkg/sigs"
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"

	yaml "gopkg.in/yaml.v3"
)

type options struct {
	dryRun               bool
	sigsFilePath         string
	labelsConfigFilePath string
	orgsConfigFilePath   string
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
	fs.StringVar(&o.sigsFilePath, "sigs_file_path", "./sigs.yaml", "DEPRECATED: file path to the sigs.yaml file to check")
	fs.StringVar(&o.sigsFilePath, "sigs-file-path", "./sigs.yaml", "file path to the sigs.yaml file to check")
	fs.StringVar(&o.labelsConfigFilePath, "labels-file-path", "../project-infra/github/ci/prow-deploy/kustom/base/configs/current/labels/labels.yaml", "file path to the labels.yaml file to check")
	fs.StringVar(&o.orgsConfigFilePath, "orgs-file-path", "../project-infra/github/ci/prow-deploy/kustom/base/configs/current/orgs/orgs.yaml", "file path to the orgs.yaml file to check")
	err := fs.Parse(os.Args[1:])
	if err != nil {
		log.Fatalf("error parsing arguments %v: %v", os.Args[1:], err)
	}
	return o
}

func init() {
	log.SetFormatter(&log.JSONFormatter{})
}

func main() {
	opts := gatherOptions()
	if err := opts.Validate(); err != nil {
		log.Fatalf("invalid arguments: %v", err)
	}
	log.Infof("dry-run: %v", opts.dryRun)

	sigsYAML, err := sigs.ReadFile(opts.sigsFilePath)
	if err != nil {
		log.Fatalf("invalid arguments: %v", err)
	}

	labelsYAML, err := labels.ReadFile(opts.labelsConfigFilePath)
	if err != nil {
		log.Fatalf("invalid arguments: %v", err)
	}

	orgsYAML, err := orgs.ReadFile(opts.orgsConfigFilePath)
	if err != nil {
		log.Fatalf("invalid arguments: %v", err)
	}

	kubevirtOrg := orgsYAML.Orgs["kubevirt"]

	for _, sig := range sigsYAML.Sigs {
		validateGroup("sig", sig, labelsYAML, kubevirtOrg)
	}

	for _, wg := range sigsYAML.Workinggroups {
		validateGroup("wg", wg, labelsYAML, kubevirtOrg)
	}

	for _, ug := range sigsYAML.Usergroups {
		validateGroup("ug", ug, labelsYAML, kubevirtOrg)
	}

	for _, committee := range sigsYAML.Committees {
		validateGroup("committee", committee, labelsYAML, kubevirtOrg)
	}

	output, err := yaml.Marshal(sigsYAML)
	if err != nil {
		log.Fatalf("in file %q: %v", opts.sigsFilePath, err)
	}
	if opts.dryRun {
		_, err := os.Stdout.Write(output)
		if err != nil {
			log.Fatalf("file %q: %v", opts.sigsFilePath, err)
		}
	} else {
		stat, err := os.Stat(opts.sigsFilePath)
		if err != nil {
			log.Fatalf("stat for file %q failed: %v", opts.sigsFilePath, err)
		}
		err = ioutil.WriteFile(opts.sigsFilePath, output, stat.Mode())
		if err != nil {
			log.Fatalf("write to file %q failed: %v", opts.sigsFilePath, err)
		}
	}

}

func validateGroup(groupType string, groupToValidate *sigs.Group, labelsYAML *labels.LabelsYAML, kubevirtOrg orgs.Org) {
	groupLog := log.WithField(groupType, groupToValidate.Name)

	// check dir exists
	if groupToValidate.Dir != "" {
		stat, err := os.Stat(groupToValidate.Dir)
		if err != nil {
			groupLog.Errorf("dir %q not found: %v", groupToValidate.Dir, err)
			groupToValidate.Dir = ""
		} else if !stat.IsDir() {
			groupLog.Errorf("dir %q is not a directory", groupToValidate.Dir)
			groupToValidate.Dir = ""
		}
	}

	// check label exists
	if groupToValidate.Label != "" {
		foundLabel := false
		for _, label := range labelsYAML.Default.Labels {
			if label.Name == groupToValidate.Label {
				foundLabel = true
				break
			}
		}
		if !foundLabel {
			groupLog.Errorf("label %q not found", groupToValidate.Label)
			groupToValidate.Label = ""
		}
	}

	// check leads - github handles are part of org
	var checkedMembers []*sigs.Lead
	for _, orgMember := range groupToValidate.Leads {
		if !kubevirtOrg.HasMember(orgMember.Github) {
			groupLog.Errorf("lead %q not found", orgMember)
		} else {
			checkedMembers = append(checkedMembers, orgMember)
		}
	}
	groupToValidate.Leads = checkedMembers

	// check chairs - github handles are part of org
	if groupToValidate.Leadership != nil {
		var checkedLeadership []*sigs.Chair
		for _, orgMember := range groupToValidate.Leadership.Chairs {
			if !kubevirtOrg.HasMember(orgMember.Github) {
				groupLog.Errorf("leadership chair %q not found", orgMember)
			} else {
				checkedLeadership = append(checkedLeadership, orgMember)
			}
		}
		groupToValidate.Leadership.Chairs = checkedLeadership
	}

	// check subprojects
	for _, subProject := range groupToValidate.SubProjects {
		subprojectLog := groupLog.WithField("subproject", subProject.Name)
		foundOwners := validateOwnersReferences(subProject, subprojectLog)
		subProject.Owners = foundOwners

		// check subproject leads - github handles are part of org
		var checkedSubprojectChairs []*sigs.Lead
		for _, orgMember := range subProject.Leads {
			if !kubevirtOrg.HasMember(orgMember.Github) {
				subprojectLog.Errorf("lead %q not found", orgMember)
			} else {
				checkedSubprojectChairs = append(checkedSubprojectChairs, orgMember)
			}
		}
		subProject.Leads = checkedSubprojectChairs

	}
}

func validateOwnersReferences(subProject *sigs.SubProject, subprojectLog *log.Entry) []string {
	foundOwners := make([]string, 0)
	for _, ownersFileURL := range subProject.Owners {
		response, err := http.DefaultClient.Head(ownersFileURL)
		if err != nil {
			subprojectLog.Errorf("failed to retrieve %q, continuing with next", ownersFileURL)
			continue
		}
		defer response.Body.Close()
		if response.StatusCode >= http.StatusOK && response.StatusCode < http.StatusMultipleChoices {
			foundOwners = append(foundOwners, ownersFileURL)
		} else {
			subprojectLog.Errorf("failed to retrieve %q: %d", ownersFileURL, response.StatusCode)
		}
	}
	return foundOwners
}
