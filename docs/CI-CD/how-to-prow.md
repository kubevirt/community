# Prow and Quickstart

Prow is a Kubernetes based CI/CD system. Jobs can be triggered by various types of events and report their status to many different services. In addition to job execution, Prow provides GitHub automation in the form of policy enforcement, chat-ops via /foo style commands, and automatic PR merging.

**NOTE|WARNING**: In order to make Prow work fine with your repo, the Kubernetes cluster **MUST** be reachable by GitHub Webhook. Then the most used option is to deploy it on GKE directly. 

## Core Components

- `hook` is the most important piece. It is a stateless server that listens for GitHub webhooks and dispatches them to the appropriate plugins. Hook's plugins are used to trigger jobs, implement 'slash' commands, post to Slack, and more. See the [`prow/plugins`](/prow/plugins/) directory for more information on plugins.
- `plank` is the controller that manages the job execution and lifecycle for jobs that run in k8s pods.
- `deck` presents a nice view of [recent jobs](https://prow.k8s.io/), [command](https://prow.k8s.io/command-help) and [plugin](https://prow.k8s.io/plugins) help information, the [current status](https://prow.k8s.io/tide) and (history)[https://prow.k8s.io/tide-history] of merge automation, and a [dashboard for PR authors](https://prow.k8s.io/pr).
- `horologium` triggers periodic jobs when necessary.
- `sinker` cleans up old jobs and pods.
- `tide` manages retesting and merging PRs once they meet the configured merge criteria. See [its README](./tide/README.md) for more information.
- `crier` manages the notifications against different providers like slack, github, etc..

Reference: https://raw.githubusercontent.com/kubernetes/test-infra/master/prow/cmd/README.md


## Deploying on

- Following **https://github.com/kubernetes/test-infra/blob/master/prow/getting_started_deploy.md**

- Deploy instance on libvirt with terraform:
```
cd ~cnv/repos/kubevirt-tutorial/administrator/terraform/libvirt
terraform init -get -upgrade=true
terraform apply -var-file varfiles/jparrill.tf -refresh=true -auto-approve
```

- Install [Golang](https://linux4one.com/how-to-install-go-on-centos-7/), [Bazel](https://docs.bazel.build/versions/master/install-redhat.html) and Tackel on guest
```
cd ~ && curl -O https://dl.google.com/go/go1.11.5.linux-amd64.tar.gz
sha256sum go1.11.5.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.11.5.linux-amd64.tar.gz

cat <<EOF > $HOME/.bash_profile \
export GOPATH=$HOME/go
export PATH=$PATH:/usr/local/go/bin:$GOPATH/bin
EOF
mkdir $HOME/go
source $HOME/.bash_profile

## Jobs by Bazel will need GCC
sudo yum groupinstall "development tools" -y
sudo yum install wget -y
sudo wget https://copr.fedorainfracloud.org/coprs/vbatts/bazel/repo/epel-7/vbatts-bazel-epel-7.repo -O /etc/yum.repos.d/bazel.repo
sudo yum install bazel -y
go get -u k8s.io/test-infra/prow/cmd/tackle
```

- Create cluster elements to work with GitHub
```
kubectl create clusterrolebinding cluster-admin-binding-kubernetes-admin --clusterrole=cluster-admin --user=kubernetes-admin
mkdir ~/private
openssl rand -hex 20 > $HOME/private/HMAC_TOKEN
kubectl create secret generic hmac-token --from-file=hmac=$HOME/private/HMAC_TOKEN
echo "f25cc009637532179fb2cdec2d888a39749ac067" > $HOME/private/OAUTH_SECRET
kubectl create secret generic oauth-token --from-file=oauth=$HOME/private/OAUTH_SECRET
```

- Spin up Prow
```
cd $HOME && git clone https://github.com/kubernetes/test-infra.git && cd $HOME/test-infra
kubectl create namespace test-pods
kubectl config set-context $(kubectl config current-context) --namespace=default
kubectl apply -f prow/cluster/starter.yaml
```

- If you are working with a local instance, just use sshuttle in order to validate the `deck` deployment
```
# Use sshuttle to access the Prow interface
sshuttle -r jparrill@192.168.1.XXX 192.168.123.0/24 -v
```

- Then access to the NodePort using your Kubeadmin node (the deck one):
```
[kubevirt@k8s-kubemaster test-infra]$ kubectl get svc
NAME         TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)          AGE
deck         NodePort    10.102.35.212   <none>        80:32494/TCP     2d21h
hook         NodePort    10.101.54.234   <none>        8888:31050/TCP   2d21h
kubernetes   ClusterIP   10.96.0.1       <none>        443/TCP          3d15h
tide         NodePort    10.100.34.208   <none>        80:31840/TCP     2d21h
```

- In order to be more standard depending if you are using GKE, you just could use an ingress with a LB. If not you could use directly the NodePorts


- Add WebHook to Github
```
# We need to update git in order to let Bazel to use "git -C ...." sentences
sudo sh -c "cat <<EOF > /etc/yum.repos.d/wandisco-git.repo
[wandisco-git]
name=Wandisco GIT Repository
baseurl=http://opensource.wandisco.com/centos/7/git/\$basearch/
enabled=1
gpgcheck=1
gpgkey=http://opensource.wandisco.com/RPM-GPG-KEY-WANdisco
EOF"
sudo rpm --import http://opensource.wandisco.com/RPM-GPG-KEY-WANdisco
sudo yum update git -y
####

go get -u k8s.io/test-infra/experiment/add-hook
bazel run //experiment/add-hook -- \
  --hmac-path=$HOME/private/HMAC_TOKEN \
  --github-token-path=$HOME/private/OAUTH_SECRET \
  --hook-url http://kubevirt-prow-0.gce.sexylinux.net:30300/hook \
  --repo the-shadowmen \
  --confirm=false
```

- Add Config to Prow you could use [this](https://github.com/the-shadowmen/nostromo/blob/master/config.yaml) as an example
- Add Plugins to Prow you could use [this](https://github.com/the-shadowmen/nostromo/blob/master/plugins.yaml) as an example
- Add Labels management to Prow you could use [this](https://github.com/the-shadowmen/nostromo/blob/master/labels.yaml) as an example

You also have a couple of make commands to work with, on [Nostromo repo](https://github.com/the-shadowmen/nostromo/blob/master/Makefile):
    - `make update-config`
    - `make update-plugins`
    - `make update-labels`

**NOTE:** Depending on the plugin you will need to add some resources to the managed repo, like [OWNERS file](https://github.com/the-shadowmen/kubevirt.github.io/blob/master/OWNERS)

- Now it's time to validate the config and upload it to Kubernetes as configmaps
```
cd $HOME/test-infra
bazel run //prow/cmd/checkconfig -- --plugin-config=$HOME/prow_conf/plugins.yaml --config-path=$HOME/prow_conf/config.yaml
kubectl create configmap plugins \
  --from-file=$HOME/prow_conf/plugins.yaml --dry-run -o yaml \
  | kubectl replace configmap plugins -f -

cd $HOME/prow_conf && kubectl create configmap config --from-file=config.yaml=config.yaml --dry-run -o yaml | kubectl replace configmap config -f -
cd $HOME/prow_conf && kubectl create configmap plugins --from-file=$HOME/prow_conf/plugins.yaml --dry-run -o yaml | kubectl replace configmap plugins -f -

cd $HOME/prow_conf && kubectl create configmap label-config --from-file=$HOME/prow_conf/labels.yaml -o yaml
```

## Pro(w)totype

We need to emulate what we're doing with Rake and TravisCI tool using Prow but adding some features. For this we will use:

- A new [GitHub organization](https://github.com/the-shadowmen)
- A fork of [Kubevirt website repository](https://github.com/the-shadowmen/kubevirt.github.io)
- A bot account with Admin permission on the Org: [Janitor](https://github.com/janitor-bot)
- Prow Stagging [Configuration repo](https://github.com/the-shadowmen/nostromo)
- New Prow CI/CD platform for stagging testing with:
    - [Deck](http://kubevirt-prow-0.gce.sexylinux.net:30000/)
    - [Hook](http://kubevirt-prow-0.gce.sexylinux.net:30300/hook)
    - [Tide](http://kubevirt-prow-0.gce.sexylinux.net:30090) 
    - GCS bucket to store the artifacts with the proper Secret and Service Account
    - The secrets to allow Prow to deal with GitHub
        - https://github.com/kubernetes/test-infra/blob/master/prow/getting_started_deploy.md#tackle-deployment
        - https://github.com/kubernetes/test-infra/blob/master/prow/docs/pr_status_setup.md
        - https://github.com/kubernetes/test-infra/blob/master/prow/getting_started_deploy.md#create-the-github-secrets

## Prow Jobs

The jobs are managed by more than one core components:

- `Plank` as a job controller, this one manage the lifecycle of the jobs
- `Horologium` triggers periodic Jobs.
- `Tide` manages retesting and merging PRs.

where you could make some kind of debugging of your jobs submission is in Tide component.

You need to upload the jobs to a config map by default called `config`. There you could put the Prow config (separated by component) and the jobs itself. In any case you could use a plugin called `config-updater` which allows you to maintain separated the config from the jobs itself.

As we shows before, this is an [example of config file](https://github.com/the-shadowmen/nostromo/blob/master/config.yaml)

In order to execute jobs you need to configure [from this point](https://github.com/kubernetes/test-infra/blob/master/prow/getting_started_deploy.md#set-namespaces-for-prowjobs-and-test-pods), GCS to store the logs and artifacts (if applies), GitHub bot account credentials to manage the organization or repo.

You have 3 kind of jobs:

- **Periodic:** This context free jobs will be executed in a pace that you decide.
- **Presubmit:** This kind of job have the context of the Repo and PR where you are working allowing you to execute the tests directly on that branch **BEFORE** the merge happens.
- **Postsubmit:** This kind of job have the context of the Repo and PR where you are working and perform the tests **AFTER** the merge happens


## Managing notifications

In the usual postsubmit and presubmit jobs there is not problems, Github will take care about the notifications but on periodic ones you need an additional component called `Crier` which will allow you to send notifications to external communication providers like Slack, Github, Gerrit, etc...

To do that, we need some things (sample Slack):
- Apply the necessary RBAC to make work Crier on K8s (located in test-infra/prow/cluster/crier_rbac.yaml)
- Slack API Token to allow crier to push notifications to Slack, supplying K8s through  a secret on the (by default) `default` namespace
- Modify the Crier deployment [adding the proper flags](https://github.com/kubernetes/test-infra/tree/master/prow/cmd/crier), in this case `--slack-workers=n` and `--slack-token-file=path-to-tokenfile`
- Modify the config configmap to add a `slack_reporter` section including the desired configuration.
- Deploy Crier component (located in test-infra/prow/cluster/crier_deployment.yaml)

## Articles to read and understand

### Introduction

- Readme file: https://github.com/kubernetes/test-infra/blob/master/prow/README.md
- Command Help: https://prow.k8s.io/command-help
- Prow Go Doc: https://godoc.org/k8s.io/test-infra/prow
- Mandatory Article to read: https://kurtmadel.com/posts/native-kubernetes-continuous-delivery/prow/
- Prow Quickstart: https://github.com/kubernetes/test-infra/blob/master/prow/getting_started_deploy.md
- Prow Images: https://github.com/kubernetes/test-infra/blob/master/prow/cmd/README.md#core-components
- Prow PR Workflow: https://raw.githubusercontent.com/kubernetes/test-infra/master/prow/docs/pr-interactions-sequence.svg?sanitize=true
- Prow Best Practices: https://github.com/kubernetes/test-infra/blob/master/prow/cmd/tide/maintainers.md#best-practices

### Prow Plugins

- Prow Plugins: https://prow.k8s.io/plugins
- Prow Code-Review process: https://github.com/kubernetes/community/blob/master/contributors/guide/owners.md#the-code-review-process

### Prow Jobs

- Prow Jobs overview: https://kurtmadel.com/posts/native-kubernetes-continuous-delivery/prow/#prow-is-a-ci-cd-job-executor
- Life of a Prow Job: https://github.com/kubernetes/test-infra/blob/master/prow/life_of_a_prow_job.md
    - Webhook Payload sample: https://github.com/kubernetes/test-infra/tree/c8829eef589a044126289cb5b4dc8e85db3ea22f/prow/cmd/phony/examples
- Prow Jobs Deep Dive:
    - https://github.com/kubernetes/test-infra/blob/master/prow/jobs.md
    - https://github.com/kubernetes/test-infra/tree/master/prow/cmd/phaino
    - https://github.com/kubernetes/test-infra/blob/master/prow/cmd/tide/config.md

### Prow SSL
- [Configure SSL](https://github.com/kubernetes/test-infra/blob/master/prow/getting_started_deploy.md#configure-ssl)
- [Cert Manager](https://github.com/jetstack/cert-manager)
- [Cert Manager Tutorial](https://github.com/jetstack/cert-manager/blob/master/docs/tutorials/acme/quick-start/index.rst)

### Others

- Another useful Article: https://kurtmadel.com/posts/native-kubernetes-continuous-delivery/native-k8s-cd/
- Kubevirt Project-Infra: https://github.com/kubevirt/project-infra
- NGINX Ingress Controller:
	- https://github.com/kubernetes/ingress-nginx
	- https://github.com/kubernetes/ingress-nginx/blob/master/docs/deploy/index.md
	- https://github.com/kubernetes/ingress-nginx/tree/master/docs/examples
