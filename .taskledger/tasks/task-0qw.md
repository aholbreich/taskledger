---
id: task-0qw
title: make repository push work like it works with adr-tool
status: open
priority: medium
created_at: 2026-05-17T20:49:54Z
updated_at: 2026-05-17T20:49:54Z
created_by: human
assignee: null
depends_on: []
claim:
  actor: null
  claimed_at: null
  expires_at: null
  heartbeat_at: null
tags: []
---

## Description

 the doc sesribes the rsource as be availabe in # Documentation: https://aholbreich.github.io/rpm-repo/#installation-fedora-centos-redhat
echo '[Holbreich]
name=Holbreich Repository
baseurl=https://aholbreich.github.io/rpm-repo/
enabled=1
gpgcheck=0' | sudo tee /etc/yum.repos.d/holbreich.repo but this is not true. the .gthub actions need to be copied form adr tool
