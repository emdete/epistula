import vmime

// see https://github.com/vlang/v/blob/master/doc/docs.md
// see https://modules.vlang.io/
import os

fn main() {
	mut session := vmime.session_open()
	defer { session.close() }
	mut email := session.email_new()
	for arg in os.args[1..] {
		if arg.starts_with("--") {
			x := arg[2..].split_nth('=', 2)
			match x[0] {
				"bcc" { email.add_bcc(x[1]) }
				"cc" { email.add_cc(x[1]) }
				"from" { email.add_from(x[1]) }
				"message-id" { email.set_message_id(x[1]) }
				"references" { email.set_references(x[1]) }
				"reply-to" { email.add_reply_to(x[1]) }
				"in-reply-to" { email.set_in_reply_to(x[1]) }
				"subject" { email.set_subject(x[1]) }
				"to" { email.add_to(x[1]) }
				"text" { email.set_text_from_file(x[1]) }
				else { }
			}
			eprintln("arg $x")
		} else {
			// to
		}
	}
	vmime.samplerun()
}

