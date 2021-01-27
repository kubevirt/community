# Title

Introducing new Kubevirt driver for Ansible Molecule 

# Abstract

Molecule is a well known test framework for Ansible. But when you run your Molecule test in Kubernetes, no real good solution exists. I'm working on creating new Molecule driver for Kubevirt to find a better approach and get a 100% pure Kubernetes solution. 

In this session I will introduce quickly why it may be better than actual drivers, how it works, and make a demo.

# Presenters

- Joël Séguillon : Senior Devops Consultant in mission at www.ateme.com - joel.seguillon@gmail.com, https://github.com/jseguillon, https://www.linkedin.com/in/jo%C3%ABl-s%C3%A9guillon-91a55814/

[X] The presenters agree to abide by the
    [Linux Foundation's Code of Conduct for Events](https://events.linuxfoundation.org/about/code-of-conduct/)

# Session details

- Track: User
- Session type: Presentation
- Duration: 20m
- Level: Any

# Additional notes

This presentation is a slideshow for overview of the driver then demo on my desk.

The project is in very early version but you can see :
- the driver, forked from molecule-docker and modified [passing tests on github actions](https://github.com/jseguillon/molecule-kubevirt/actions/runs/502172219)
- this github fork from geerlingguy role showing [its working for Centos7](https://github.com/jseguillon/ansible-role-nginx/actions/runs/502378610)
