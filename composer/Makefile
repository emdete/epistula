#!/usr/bin/env make -f
VFLAGS=-show-c-output -showcc -g -keepc -cg
VFLAGS=-keepc -cg
TEMPFILE=/tmp/epistula-composer-tmp

all: epistula-composer

run: epistula-composer
	echo € > $(TEMPFILE)
	./$< \
			--bcc="M. <mdt@emdete.de>" \
			--cc="M. <mdt@emdete.de>" \
			--from="M. <mdt@emdete.de>" \
			--reply \
			--reply-message-id=000 \
			--reply-text=$(TEMPFILE) \
			--subject=Epistula \
			--text=/etc/hostname \
			--to="M. <mdt@emdete.de>" \
			--pid=0 \
		> /tmp/epistula-composer

dbg:
	make -C .. dbg

epistula-composer: epistula-composer.v gmime/*.v
	v $(VFLAGS) epistula-composer.v

uitest: uitest.v
	v $(VFLAGS) uitest.v

up:
	v up

win:
	v -o epistula-composer.exe -os windows epistula-composer.v

and:
	v -o epistula-composer.apk -os android epistula-composer.v

