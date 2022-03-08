import vmime

// see https://github.com/vlang/v/blob/master/doc/docs.md
// see https://modules.vlang.io/
import os

fn main() {
	mut session := vmime.session_open()
	defer { session.close() }
	mut email := session.email_new()
	defer { email.close() }
	for arg in os.args[1..] {
		if arg.starts_with("--") {
			x := arg[2..].split_nth('=', 2)
			match x[0] {
				"bcc" { email.add_bcc(x[1]) }
				"cc" { email.add_cc(x[1]) }
				"from" { email.add_from(x[1]) }
				"message-id" { email.set_in_reply_to(x[1]) }
				"references" { email.set_references(x[1]) }
				"reply-to" { email.add_reply_to(x[1]) }
				"in-reply-to" { x[1] }
				"subject" { email.set_subject(x[1]) }
				"to" { email.add_to(x[1]) }
				"text" { email.set_text(read_file(x[1])) }
				else { }
			}
			eprintln("arg $x")
		} else {
			// to
		}
	}
	// "X-Epistula-Status", "I am not done"
	email.edit()
	email.set_user_agent("Epistula")
	email.set_date_now()
	email.set_message_id("epistula.de")
}

fn read_file(filename string) string {
	mut buffer := []byte{}
	buffer = os.read_bytes(filename) or { panic(err) }
	return buffer.bytestr()
}

//	multipart := C.g_mime_multipart_new_with_subtype(cstr("mixed"))
//	C.g_mime_text_part_set_charset(textpart, cstr("utf-8"))
//	C.g_mime_multipart_add(multipart, C.GMIME_OBJECT(textpart))
//	mail_attach(multipart, "../screenshot.png")
//	C.g_object_unref(C.G_OBJECT(multipart))
//	mmsg := email_parse(tempfile)
//	status := unsafe { C.g_mime_object_get_header(C.GMIME_OBJECT(mmsg), cstr("X-Epistula-Status")).vstring() }
//	mail_walk(mmsg, fn (part &C._GMimeObject) bool {
//		ct := C.g_mime_object_get_content_type (C.GMIME_OBJECT(part))
//		s := unsafe { C.g_mime_content_type_get_mime_type (ct).vstring() }
//		return true
//	})
