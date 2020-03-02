**Table of contents**

<!-- TOC depthFrom:1 insertAnchor:false orderedList:true updateOnSave:true -->

- [Video Recording](#video-recording)
- [Audio recording](#audio-recording)
- [Video assembling](#video-assembling)

<!-- /TOC -->

# Video Recording

- We use OBS for recording one screen in maximum size at 24 fps and export as MP4 file, recording one file per item to show, in order to reduce the amount of re-recording needed if something goes wrong. We don't care about audio here as we'll be recording audio at a later point

# Audio recording

- Audacity is used to record the audio snippets while watching the videos recorded in prior step, so that audio lenght is more or less similar to video.
- For post processing, an area without 'talk' is selected and `Effect/Noise Reduction/Create Noise profile` is selected, then the full track (`CTRL-A`) is selected, and `Effect/Noise Reduction/Accept` is executed to perform the noise reduction on the whole track

# Video assembling

- kdenlive is used to assemble audio tracks and video tracks
- Additional `in` and `out` sequences are used for the logo
- For titles, a title slide is used with fade-in with the text of the section
  - A background in `KubeVirt green` is used so that it doesn't happen over a black background
