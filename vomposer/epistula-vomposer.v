import vmime

// see https://github.com/vlang/v/blob/master/doc/docs.md
// see https://modules.vlang.io/
import os

fn main() {
	mut session := vmime.session_open()
	defer { session.close() }
	mut to_list := session.address_list_new()
	mut cc_list := session.address_list_new()
	mut bcc_list := session.address_list_new()
	mut from_list := session.address_list_new()
	mut sender_list := session.address_list_new()
	mut reply_to_list := session.address_list_new()
	mut attachment_list := []string{}
	mut in_reply_to := ""
	mut references := ""
	mut subject := ""
	mut text := ""
	mut pid := 0
	for arg in os.args[1..] {
		if arg.starts_with("--") {
			x := arg[2..].split_nth('=', 2)
			match x[0] {
				"to" { to_list.add(x[1]) }
				"cc" { cc_list.add(x[1]) }
				"bcc" { bcc_list.add(x[1]) }
				"from" { from_list.add(x[1]) }
				"reply-to" { reply_to_list.add(x[1]) }
				"pid" { pid = x[1].int() }
				"sender" { sender_list.add(x[1]) }
				"message-id" { in_reply_to = x[1] }
				"references" { references = x[1] }
				"subject" { subject = x[1] }
				"text" { text = read_file(x[1]) }
				"attachment" { attachment_list << x[1] }
				else { eprintln("arg $x") }
			}
		} else {
			to_list.add(arg)
		}
	}
	mut done := false
	mut abort := false
	for ! done && ! abort {
		mut edit_mail := session.email_new()
		defer { edit_mail.close() }
		edit_mail.add_from(from_list)
		edit_mail.add_to(to_list)
		edit_mail.add_cc(cc_list)
		edit_mail.add_bcc(bcc_list)
		edit_mail.set_subject(subject)
		if attachment_list.len == 0 {
			edit_mail.set_header_x("X-Epistula-Attachments", "#put space delimted list of filenames here#")
		} else {
			edit_mail.set_header_x("X-Epistula-Attachments", attachment_list.join(" "))
		}
		edit_mail.set_header_x("X-Epistula-Status", "I am not done")
		edit_mail.set_header_x("X-Epistula-Comment", "This is your MUA talking to you. Add attachments as headerfield like below. Dont destroy the mail structure, if the outcome cant be parsed you will thrown into your editor again to fix it. Change the Status to not contain 'not'. Add a 'abort' to abort sending (editings lost).")
		edit_mail.set_text(text)
		edit_mail.edit()

		status := edit_mail.get_header("X-Epistula-Status")
		abort = (status.index("abort") or { -1 }) >= 0
		if ! abort {
			attachment_list.clear()
			attachments := edit_mail.get_header("X-Epistula-Attachments")
			if ! attachments.starts_with("#") {
				for attachment in attachments.split(" ") {
					eprintln("attachment $attachment")
					attachment_list << attachment
				}
			}
			text = edit_mail.get_text()
			subject = edit_mail.get_subject()
			from_list.set(edit_mail.get_from())
			to_list.set(edit_mail.get_to())
			cc_list.set(edit_mail.get_cc())
			bcc_list.set(edit_mail.get_bcc())
			done = (status.index("not done") or { -1 }) < 0
		}
	}

	if ! abort {
		mut email := session.email_new()
		defer { email.close() }
		if from_list.len() > 0 { email.add_from(from_list) }
		if to_list.len() > 0 { email.add_to(to_list) }
		if cc_list.len() > 0 { email.add_cc(cc_list) }
		if bcc_list.len() > 0 { email.add_bcc(bcc_list) }
		if sender_list.len() > 0 { email.add_sender(sender_list) }
		if reply_to_list.len() > 0 { email.add_reply_to(reply_to_list) }
		email.set_user_agent("Epistula")
		email.set_date_now()
		email.set_message_id("epistula.de")
		if in_reply_to != "" { email.set_in_reply_to(in_reply_to) }
		if references != "" { email.set_references(references) }
		email.set_subject(subject)
		email.set_text(text)
		for attachment in attachment_list {
			email.attach(attachment)
		}
		//email.encrypt()
		email.transfer()
		if pid > 0 {
			sig := int(os.Signal.usr1)
			eprintln("kill -s $sig $pid")
			if C.kill(pid, sig) != 0 {
				eprintln("error sending signal")
			}
		}
	} else {
		eprintln("aborted")
	}
}

fn read_file(filename string) string {
	mut buffer := []u8{}
	buffer = os.read_bytes(filename) or { panic(err) }
	return buffer.bytestr()
}

