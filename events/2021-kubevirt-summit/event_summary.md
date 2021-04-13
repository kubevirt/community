# Event Summary for KubeVirt Summit

Author: Josh Berkus 

TL;DR: a successful first contributor summit for the KubeVirt project.

## Background

KubeVirt Summit was a two-day free online event, held on February 9th
and 10th 2021.  The first summit for the KubeVirt community, the event
was hosted by the CNCF and run by Red Hat staff.  The event was aimed at
KubeVirt contributors and power users.

KubeVirt is technology for running virtual machines using Kubernetes.
It was created at Red Hat, promoted as an upstream open source project,
and contributed to the CNCF in 2020.  While initially the project faced
a great deal of competition, today it's the front-runner for combining
Kubernetes and virtualization.  KubeVirt is part of the CNV and KNI
products for OpenShift, as well as a key part of our OpenShift migration
services.

The Summit was designed as a noncommercial, "hacker" event that focused
entirely on open source technology.  There were no booths or sponsors.
Homepage: https://kubevirt.io/summit/

## Attendees

The Summit was advertised only to the KubeVirt community.  It attracted
317 registrants, of which 192 attended one or more sessions the first
day, and 130 attended the second day.  Our targets for the event were
around half those amounts.

The largest group of attendees were project contributors (40%), with an
equal number split between development and production users.
Overwhelmingly attendees came from the tech industry, 91% working for
either an IT vendor or a Cloud services company.  One-quarter of
attendees worked for Red Hat or IBM, and another 5% worked for nVidia,
with the remaining 70% being a general audience.

In an online event, we were not in a position to collect diversity
statistics.

## Sessions

The Summit consisted of 20 half-hour sessions on a variety of topics,
from migrating VMs into KubeVirt, to project code management, to volumes
and network devices.  A key session was the discussion of releasing
KubeVirt version 1.0.  There were specifically no introductory sessions
at the event, which was expected to attract only people who were already
familiar with KubeVirt, and were looking for new and interesting uses of
the technology.

Speakers were primarily from Red Hat and nVidia, who are also the
largest corporate sponsors of the project.  Additional speakers included
folks from CloudBase, Intel, and SuSE.

Sessions covered 5 hours each day, starting at 14:00 UTC, in order to
attract a European and American audience.  As with other online events,
peak attendance for each day was in either the first or second session,
slowly decreasing throughout the day, from 102 for the Welcome session
down to 54 22:30 UTC. This is a common pattern for online events.

All of these sessions are available on KubeVirt Youtube:

## Virtual Conference Platform

Because it was a no-budget CNCF event, we were required to use their
contracted meetup platform, Bevy.  Bevy primarily functions as a
scheduling + streaming platform for meetups, and all CNCF-sponsored
meetups are scheduled using it.  This event was the test run for a new
feature called Bevy Events, which was supposed to support longer,
single-track conference events.  Bevy has a more expensive sell-up
option for more serious conferences.

Bevy Events, unsurprisingly, felt like holding a very long meetup.
Sessions were presented continuously in the same channel, which had some
advantages for attendance.  Like other streaming platforms, Bevy has a
chat alongside the video, which was very limited, kind of like Zoom.  We
found the Q&A functional unusable and turned it off.

On the other hand, video streaming was really solid, with few glitches
through the event, and the Bevy app being entirely in-browser limited
required troubleshooting.  The primary failure of the platform was that,
while it will allow you to schedule a 2-day event, it won't actually run
one.  We found out the hard way that you have to schedule each day as a
separate event.

As such, I would recommend this platform only for one-day, one-track
events that you are conducting for the Linux Foundation or another Bevy
customer.  It's not terrible, but there are better options available.

## Other Notes

In order to get attendee data, we ran a survey, with a drawing for
t-shirts as incentive.  In order to prevent a flood of garbage data for
a free registration event, we only advertised this survey on the event
chat.  This mostly limited survey respondees to people who had actually
attended, and I would recommend that approach for other free online events.

Nevertheless, we did get 9 out of 43 survey respondees giving us
obviously fake data.  Particularly, they evaluated sessions that hadn't
actually happened yet according to the form timestamp.  So if you do an
incentivized survey, make sure to filter it for bad data before running
charts.

## Summary

The KubeVirt summit can be considered a success by any measure of a
contributor event.  Feedback was positive, and we had double the
attendance -- and much higher "outside" attendance -- than we were
expecting.  Futher, it let us bring together the rapidly growing
KubeVirt community as a whole, something we would have had difficulty
doing at an in-person event.

The only sad thing was that we couldn't do a contributor dinner.

-- 
--
Josh Berkus
Kubernetes Community
Red Hat OSPO
