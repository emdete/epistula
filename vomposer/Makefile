#!/usr/bin/env make -f
CFLAGS=`pkg-config --cflags gmime-3.0`
LDFLAGS=`pkg-config --libs gmime-3.0`
VFLAGS=-show-c-output -showcc -g -keepc -cg
VFLAGS=-keepc -cg

all: cmime
	LANG=C ./cmime
	./cmime > /tmp/cmime

run: vmime
	./$< > /tmp/vmime

dbg:

vmime: vmime.v
	v $(VFLAGS) vmime.v

uitest: uitest.v
	v $(VFLAGS) uitest.v

cmime.o: cmime.c

cmime: cmime.o
	gcc -o cmime cmime.o $(LDFLAGS)

