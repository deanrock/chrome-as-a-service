FROM ubuntu:latest

ENV DEBIAN_FRONTEND noninteractive

RUN apt-get update

# chrome-browser
RUN apt-get install -y ca-certificates wget
RUN wget https://dl.google.com/linux/direct/google-chrome-stable_current_amd64.deb -P /tmp/
RUN dpkg -i /tmp/google-chrome-stable_current_amd64.deb || true
RUN apt-get install -fy

# pulseaudio
RUN apt-get install -y pulseaudio

# vnc server
RUN apt-get install -y tightvncserver

# SSH server for debugging
RUN apt-get install -y openssh-server && \
	mkdir /var/run/sshd && \
	chmod 0755 /var/run/sshd

# add user
RUN adduser --disabled-password --gecos "Chrome User" --uid 2000 chrome

# copy auth key & create required folders
RUN mkdir /home/chrome/.ssh && \
	mkdir /tools/
ADD id_rsa.pub /home/chrome/.ssh/authorized_keys
ADD start.sh /tools/start.sh
ADD chrome.sh /tools/chrome.sh
RUN chown -R chrome:chrome /home/chrome/.ssh && \
	apt-get install -y sudo

ADD id_rsa.pub /root/.ssh/authorized_keys
RUN chown -R root:root /root/.ssh

RUN /usr/sbin/sshd

USER chrome

RUN mkdir /home/chrome/.vnc && \
	echo 'mypass' | vncpasswd -f > /home/chrome/.vnc/passwd && \
	chmod 600 /home/chrome/.vnc/passwd && \
	echo '#!/bin/bash\n' > /home/chrome/.xsession

USER root

ENTRYPOINT ["/tools/start.sh"]

EXPOSE 22



## install pulseaudio pulseaudio-utils pavumeter pavucontrol paman paprefs pasystray
