module vmime

// a session to ensure g_mime_init/g_mime_shutdown
[heap] pub struct Session {
mut:
	open bool
}

pub fn session_open() &Session {
	C.g_mime_init()
	return &Session{ true }
}

pub fn (mut this Session) close() {
	C.g_mime_shutdown()
	this.open = false
}

