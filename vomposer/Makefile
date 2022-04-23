#!/usr/bin/env make -f
VFLAGS=-show-c-output -showcc -g -keepc -cg
VFLAGS=-keepc -cg
TEMPFILE=/tmp/epistula-vomposer-tmp

all: epistula-vomposer

run: epistula-vomposer

dbg: epistula-vomposer
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
		> /tmp/epistula-vomposer

epistula-vomposer: epistula-vomposer.v vmime/*.v
	v $(VFLAGS) epistula-vomposer.v

uitest: uitest.v
	v $(VFLAGS) uitest.v

win:
	v -o epistula-vomposer.exe -os windows epistula-vomposer.v

and:
	v -o epistula-vomposer.apk -os android epistula-vomposer.v

