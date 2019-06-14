# Prow Jobs

## Hands on

- **Following**: https://github.com/kubernetes/test-infra/blob/master/prow/getting_started_deploy.md

- [Configure Gcloud Storage](https://github.com/kubernetes/test-infra/blob/master/prow/getting_started_deploy.md#configure-cloud-storage)
```
mkdir ~/downloads
wget https://dl.google.com/dl/cloudsdk/channels/rapid/downloads/google-cloud-sdk-231.0.0-linux-x86_64.tar.gz -O google-cloud-sdk-231.0.0-linux-x86_64.tar.gz
cd downloads && tar xvzf google-cloud-sdk-231.0.0-linux-x86_64.tar.gz
cd google-cloud-sdk && ./install.sh

## Configure new SA for Prow
gcloud beta iam service-accounts create sa-comm-prow --description "SA for Prow Purposes on Communitty dept" --display-name "SA-Prow"

## Create the Bucket
## Expose it
## Grant write access to allow Prow to upload the content

## Serialize the new SA using a new key
gcloud iam service-accounts keys create ~/private/service-account.json --iam-account sa-comm-prow@cnvlab-209908.iam.gserviceaccount.com

## Create Secret file with the SA json file
kubectl create secret generic gcs --from-file=service-account.json

## Check prow config and Submit
## Use the nostromo repository to find the config.yaml and plugins.yaml files
run --verbose_failures //prow/cmd/checkconfig -- \
    --plugin-config=/home/jparrill/ownCloud/RedHat/RedHat_Engineering/kubevirt/CI-CD/Prow/repos/nostromo/plugins.yaml \
    --config-path=/home/jparrill/ownCloud/RedHat/RedHat_Engineering/kubevirt/CI-CD/Prow/repos/nostromo/config.yaml

# Ensure that the new config are loaded before execute any other action, you could check it just seeing the Tide log
```

## References

- Prow Jobs overview: https://kurtmadel.com/posts/native-kubernetes-continuous-delivery/prow/#prow-is-a-ci-cd-job-executor
- Life of a Prow Job: https://github.com/kubernetes/test-infra/blob/master/prow/life_of_a_prow_job.md
    - Webhook Payload sample: https://github.com/kubernetes/test-infra/tree/c8829eef589a044126289cb5b4dc8e85db3ea22f/prow/cmd/phony/examples
- Prow Jobs Deep Dive:
    - https://github.com/kubernetes/test-infra/blob/master/prow/jobs.md
    - https://github.com/kubernetes/test-infra/tree/master/prow/cmd/phaino
    - https://github.com/kubernetes/test-infra/blob/master/prow/cmd/tide/config.md
