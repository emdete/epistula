module gmime

import log

// a session to ensure g_mime_init/g_mime_shutdown
[heap] pub struct Session {
mut:
	open bool
	logger log.Log
}

pub fn session_open(logger log.Log) &Session {
	C.g_mime_init()
	return &Session{
		true
		logger
	}
}

pub fn (mut this Session) close() {
	C.g_mime_shutdown()
	this.open = false
}

