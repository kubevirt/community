
# Overview
This design provides an approach to allow DataImportCron to import HTTP sources.  
The design will also demonstrate to the user how to operate with such import types.  

## Motivation
Currently, the DataImportCron allows the import of Registry Imports only. Recently, there was a demand to also allow HTTP import types.  
The problem that arises from HTTP imports is that there is no convention between the different sources, so it's hard to know when the image is updated for each source in a generic way, which will make the polling process more difficult than standard registry sources.  
One possible solution If the URL is static is to use a Get request with an If-Modified-Since header where we specify the date from which we want to check if there was a change. If there was a change since the specified date, the request will return with a status of 200OK, and then we know that the image has been updated since that date.  

In the current situation we can enable HTTP imports by updating the DesiredDigest label.  

## Goals

* Allow the user to perform http imports manually with dataimportcron
* Create a poller that will cover as many cases as possible for automatic updating

## Non Goals

* The poller will probably not cover all import cases and sometimes the user will have to manually update the digest

## User Stories

* As a user, I would like to import images from an HTTP source using DataImportCron
* As a user, I would like the poller automatically update the image when it is needed
* As a user, I would like to manually trigger an HTTP import with DataImportCron

## Repos

* **CDI**: Offers an API which serves as a mutable symbolic link to a Persistent Volume Claim (PVC). This link can be used in a DataVolume.
* **CDI**: Provides an API called DataImportCron, designed to manage the periodic polling and importing of disk images as PVCs into a specific namespace (golden images by default). This functionality leverages the CDI import processes.

## Implementation Phases

**Phase1** - Manual Import Update
* DataImportCron controller to handle HTTP imports
* DataImportCron webhook/validation to support HTTP source type

**Phase2** - Poller
* DataImportCron poller to support HTTP source polling

# Design

If the URL is static:
* Make a GET request with the If-Modified-Since Header starting from the date stored in the AnnLastCronTime annotation
* If the returned status is 200OK, perform import again
* Update AnnLastCronTime to time.Now()

## DataImportCron Example

This example polls http://tinycorelinux.net/14.x/x86/release/Core-14.0.iso static image. The DataImportCron will import it as new PVCs and automatically manage updating the corresponding DataSource.

* **source** specifies where to poll from
* **http** specifies that the type is HTTP import
* **url** specifies the image URL
* **schedule** should be empty when we are manually updating the DataImportCron desired digest annotation

```yaml
apiVersion: cdi.kubevirt.io/v1beta1
kind: DataImportCron
metadata:
  name: fedora-image-import-cron
  namespace: golden-images
spec:
  template:
    spec:
      source:
        http:
          url: "http://tinycorelinux.net/14.x/x86/release/Core-14.0.iso"
      storage:
        resources:
          requests:
            storage: 5Gi
  schedule: ""
  garbageCollect: Outdated
  importsToKeep: 2
  managedDataSource: fedora
```
