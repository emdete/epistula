import vmime
import notmuchconfig

// see https://github.com/vlang/v/blob/master/doc/docs.md
// see https://modules.vlang.io/
import os

fn main() {
	config := notmuchconfig.new_config()
	mut session := vmime.session_open()
	defer { session.close() }
	mut attachment_list := []string{}
	mut bcc_list := session.address_list_new()
	mut cc_list := session.address_list_new()
	mut from_list := session.address_list_new()
	mut in_reply_to := ""
	mut references := ""
	mut reply_to := ""
	mut subject := ""
	mut text := ""
	mut to_list := session.address_list_new()
	mut pid := 0
	from_list.add(config.user_name + " <" + config.user_primary_email + ">")
	for arg in os.args[1..] {
		if arg.starts_with("--") {
			x := arg[2..].split_nth('=', 2)
			match x[0] {
				"to" {
					to_list.add(x[1])
				}
				"cc" {
					cc_list.add(x[1])
				}
				"bcc" {
					bcc_list.add(x[1])
				}
				"from" {
					to_list.add(x[1])
				}
				"reply-to" {
					reply_to = x[1]
					to_list.add(x[1])
				}
				"pid" { pid = x[1].int() }
				"message-id" { in_reply_to = x[1] }
				"references" { references = x[1] }
				"subject" { subject = x[1] }
				"text" {
					text = read_file(x[1])
					text = "> " + text.replace("\n", "\n> ")
				}
				"attachment" { attachment_list << x[1] }
				else { eprintln("unknown arg $x") }
			}
		} else {
			to_list.add(arg)
		}
	}
	title := "Epistula Composer: " + config.user_name + " <" + config.user_primary_email + ">" + " to " + reply_to
	mut stdout := os.stdout()
	stdout.write(("\x1b]1;"+title+"\a\x1b]2;"+title+"\a").bytes())?
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
		edit_mail.set_text(text, true)
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
		email.add_bcc(bcc_list)
		email.add_cc(cc_list)
		email.add_from(from_list)
		email.add_reply_to(from_list)
		email.add_to(to_list)
		email.set_user_agent("Epistula")
		email.set_date_now()
		email.set_message_id("epistula.de")
		email.set_in_reply_to(in_reply_to)
		email.set_references(references)
		email.set_subject(subject)
		email.set_text(text, false)
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

