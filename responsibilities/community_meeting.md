# Community Meeting

At KubeVirt's present scale the community opts to host a weekly meetings for all interested contributors.  This gives contributors another channel to discuss important topics.  There is also value in adding a `Let's get to know each other` touch to project work.

## Meeting Mechanics

* Where: Zoom meeting ID: 92221936273
  * Link to join is [here](https://zoom.us/j/92221936273)
  * Community organizers will use CNCF/KubeVirt account to host meeting

* When: Every Wednesday @ 16:00 CET/CEST (10:00 EST/EDT) calendar event.  See comm cal [here](https://calendar.google.com/calendar/u/0/embed?src=18pc0jur01k8f2cccvn5j04j1g@group.calendar.google.com&ctz=Etc/GMT)
  * Community organizers will start meeting approx 10 mins before start time
  * Community organizers will start recording and begin session at 03 mins past the hour
  * Community organizers will lead the community through the agenda

* Meeting hosts: @ccallegar, @sgott

* Transcription
  * On going meeting notes file: [here](https://docs.google.com/document/d/1kyhpWlEPzZtQJSjJlAqhPcn3t0Mt_o0amhpuNPGs1Ls)
  * Community organizers will attempt to transcribe each topic in the agenda

* Recordings
  * Community organizers will process Zoom mp4 digital audio/video recording into a recording using Open Source codecs [VP9](https://en.wikipedia.org/wiki/VP9) video and [Opus](https://en.wikipedia.org/wiki/Opus_%28audio_format%29) audio
```
  cd zoom/DATE
  ffmpeg -i "${I}" -c:v libvpx-vp9 -c:a libopus $(basename "`pwd`" | awk {'print $1 "_" $3 "_" $4 "_" $5'}).webm
```
  * Community organizers will upload the video to KubeVirt’s YouTube channel and the Community Meetings playlist.
  * Community organizers will add the YouTube link to the meeting notes

* Meeting minutes are sent to the kubevirt-dev mailing list/group:
  * mail-to: kubevirt-dev@googlegroups.com
  * Subject: KubeVirt Weekly Community Meeting Minutes DATE


## Meeting Content

#### Introductions
* Community organizers should introduce themselves and the meeting
* Community organizers should allow new attendees to introduce themselves

#### Agenda - typically planned discussions added throughout the week
* Community organizers should introduce each topic contributor and the topic
* Community organizers should allow ample time to discuss the topic.  Some topics may need to be time boxed or broken out into mailing list, slack and/or separate community meetings

#### Open Agenda - typically ad hoc discussions
* Community organizers should introduce each topic contributor and the topic
* Community organizers should allow ample time to discuss the topic.  Some topics may need to be time boxed or broken out into mailing list, slack and/or separate community meetings

#### Pull Requests
* Community organizers should introduce each pull request contributor and the pull request
* Community organizers should allow ample time to discuss the pull req.  Some topics may need to be time boxed or broken out into mailing list, slack and/or separate community meetings

#### Bug Scrub - typically engs will run this section and review unreviewed issues
Last 20 minutes of each community meeting are dedicated to scrubbing of KubeVirt Issues. Precedence should be given to Issues raised by attendees. The remaining of the time should be spent going over Issues one by one starting from the most recent one. Each Issue we visit should be given one of the following triage labels:

* triage/accepted: Indicates an issue or PR is ready to be actively worked on.
* triage/duplicate: Indicates an issue is a duplicate of other open issue.
* triage/needs-information: Indicates an issue needs more information in order to work on it.
* triage/not-reproducible: Indicates an issue can not be reproduced as described.
* triage/unresolved: Indicates an issue that can not or will not be resolved.

Some topics may need to be time boxed or broken out into mailing list, slack and/or separate community meetings

Repos that should be reviewed:
* [kubevirt/kubevirt](https://github.com/kubevirt/kubevirt/issues)
* [kubevirt/user-guide](https://github.com/kubevirt/user-guide/issues)
