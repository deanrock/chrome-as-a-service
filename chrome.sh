#!/bin/bash
set -e

vncserver
DISPLAY=:1 google-chrome --no-sandbox
