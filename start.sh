#!/bin/bash
set -e

/usr/sbin/sshd

su -c "/tools/chrome.sh" chrome

/bin/bash