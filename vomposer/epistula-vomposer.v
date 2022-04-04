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
			email.add_to(arg)
		}
	}
	mut done := false
	mut abort := false
	for ! done && ! abort{
		email.set_header_x("X-Epistula-Status", "I am not done")
		email.set_header_x("X-Epistula-Comment", "This is your MUA talking to you. Add attachments as headerfield like below. Dont destroy the mail structure, if the outcome cant be parsed you will thrown into your editor again to fix it. Change the Status to not contain 'not'. Add a 'abort' to abort sending (editings lost).")
		email.set_header_x("X-Epistula-Attachments", "#put space delimted list of filenames here#")
		email.edit()

		status := email.get_header("X-Epistula-Status")
		done = (status.index("not done") or { -1 }) < 0
		abort = (status.index("abort") or { -1 }) >= 0

		attachments := email.get_header("X-Epistula-Attachments")
		if ! attachments.starts_with("#") {
			for attachment in attachments.split_any(" ;,") {
				eprintln("attachment $attachment")
				email.attach(attachment)
			}
		}
	}

	if ! abort {
		email.set_user_agent("Epistula")
		email.set_date_now()
		email.set_message_id("epistula.de")
		email.encrypt()
	} else {
		eprintln("aborted")
	}
}

fn read_file(filename string) string {
	mut buffer := []byte{}
	buffer = os.read_bytes(filename) or { panic(err) }
	return buffer.bytestr()
}

