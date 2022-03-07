#!/usr/bin/env make -f
CFLAGS=`pkg-config --cflags gmime-3.0`
LDFLAGS=`pkg-config --libs gmime-3.0`
VFLAGS=-show-c-output -showcc -g -keepc -cg
VFLAGS=-keepc -cg
TEMPFILE=/tmp/epistula-vomposer-tmp

all: cmime
	./$< \
		> /tmp/cmime

run: epistula-vomposer
	./$< \
			--from="M. <mdt@emdete.de>" \
			--reply \
			--reply-text=$(TEMPFILE) \
			--reply-message-id=000 \
			--to="M. <mdt@emdete.de>" \
			--subject=Epistula \
			--cc="M. <mdt@emdete.de>" \
			--bcc="M. <mdt@emdete.de>" \
		> /tmp/epistula-vomposer

dbg:

epistula-vomposer: epistula-vomposer.v
	v $(VFLAGS) epistula-vomposer.v

uitest: uitest.v
	v $(VFLAGS) uitest.v

cmime.o: cmime.c

cmime: cmime.o
	gcc -o cmime cmime.o $(LDFLAGS)

