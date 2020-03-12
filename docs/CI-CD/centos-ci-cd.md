# Centos ci-cd

## Introduction

In some of the KubeVirt repositories the ci-cd automation is done through the CentOS ci-cd located in https://ci.centos.org

That site uses the authentication sidecar of the CentOS OpenShift system in https://console.apps.ci.centos.org:8443

You need to get an account in that system to be able to deploy new ci-cd flows or check details of the already existing implementation.

## How to get an account

- The official guide is written in: https://wiki.centos.org/SIGGuide#head-593a859807e26a8299efeebda9e0aaf0d0b3a821

To short it basically you have to:

1. Visit https://bugs.centos.org - Create your account in the sign-up page https://bugs.centos.org/signup_page.php - Confirm your account with the email you'll get - Login within your new credentials
2. Report an Issue under the 'CentOS CI' project
   - Include the following information in your report (an example is included in the references point):
     - Your Name
     - The project you are working with
     - Your Desired Username
     - Your Email Address
     - Your GPG Public Key (attach this to the bug please) (check references for how-to documents and tips)

## References

- Official guide: https://wiki.centos.org/SIGGuide#head-593a859807e26a8299efeebda9e0aaf0d0b3a821
- Issue template/example: https://bugs.centos.org/view.php?id=16294
- CentOS ci-cd: https://ci.centos.org
- CentOS bug website: https://bugs.centos.org
- GPG interesting links:
  - Fedora Project wiki - Creating GPG Keys: https://fedoraproject.org/wiki/Creating_GPG_Keys
  - GitHub how-to generating a new GPG key: https://help.github.com/en/articles/generating-a-new-gpg-key
  - Best practices: https://riseup.net/en/security/message-security/openpgp/best-practices
  - Best practices: https://blog.mailfence.com/openpgp-encryption-best-practices/
  - Backup tips: https://msol.io/blog/tech/back-up-your-pgp-keys-with-gpg/
