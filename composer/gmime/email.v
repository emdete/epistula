module gmime

import io.util
import os
import log

// representing an email
pub struct Email {
mut:
	message &C._GMimeMessage // gmime3 mail message structure
	multipart &C._GMimeMultipart // multipart content
	logger log.Log
}

// create new from session
pub fn (this &Session) email_new() &Email {
	message := C.g_mime_message_new(C.gboolean(1))
	multipart := C.g_mime_multipart_new_with_subtype(cstr("mixed"))
	C.g_mime_message_set_mime_part(message, C.GMIME_OBJECT(multipart))
	//this.logger.debug("new message $message")
	return &Email{
		message
		multipart
		this.logger
	}
}

// free mem
pub fn (mut this Email) close() {
	this.logger.debug("close message $this.message")
	C.g_object_unref(C.G_OBJECT(this.message))
	C.g_object_unref(C.G_OBJECT(this.multipart))

}

// parse file
pub fn (mut this Email) parse(filename string) {
	err := &C._GError(0)
	stream := C.g_mime_stream_fs_open(cstr(filename), /*O_RDONLY*/0, 0644, &err)
	if stream == voidptr(0) { panic( vstr(err.message) ) }
	defer { C.g_object_unref(C.G_OBJECT(stream)) }
	parser := C.g_mime_parser_new_with_stream(stream)
	defer { C.g_object_unref(C.G_OBJECT(parser)) }
	C.g_object_unref(C.G_OBJECT(this.message))
	this.message = C.g_mime_parser_construct_message(parser, /*NULL*/voidptr(0))
	this.multipart = C.g_mime_multipart_new_with_subtype(cstr("mixed")) // TODO: fill text in?
}

// set to
pub fn (mut this Email) add_to(value &AddressList) {
	value.iterate(fn [mut this](fullname string, emailaddress string, cset string) {
		C.g_mime_message_add_mailbox(this.message, C.GMimeAddressType(C.GMIME_ADDRESS_TYPE_TO), cstr(fullname), cstr(emailaddress))
	})
}

pub fn (mut this Email) get_to() &C._InternetAddressList {
	return C.g_mime_message_get_to (this.message)
}

// set carbon copy
pub fn (mut this Email) add_cc(value &AddressList) {
	value.iterate(fn [mut this](fullname string, emailaddress string, cset string) {
		C.g_mime_message_add_mailbox(this.message, C.GMimeAddressType(C.GMIME_ADDRESS_TYPE_CC), cstr(fullname), cstr(emailaddress))
	})
}

pub fn (mut this Email) get_cc() &C._InternetAddressList {
	return C.g_mime_message_get_cc (this.message)
}

// set blind carbon copy
pub fn (mut this Email) add_bcc(value &AddressList) {
	value.iterate(fn [mut this](fullname string, emailaddress string, cset string) {
		C.g_mime_message_add_mailbox(this.message, C.GMimeAddressType(C.GMIME_ADDRESS_TYPE_BCC), cstr(fullname), cstr(emailaddress))
	})
}

pub fn (mut this Email) get_bcc() &C._InternetAddressList {
	return C.g_mime_message_get_bcc (this.message)
}

// set from
pub fn (mut this Email) add_from(value &AddressList) {
	value.iterate(fn [this](fullname string, emailaddress string, cset string) {
		C.g_mime_message_add_mailbox(this.message, C.GMimeAddressType(C.GMIME_ADDRESS_TYPE_FROM), cstr(fullname), cstr(emailaddress))
	})
}

pub fn (mut this Email) get_from() &C._InternetAddressList {
	return C.g_mime_message_get_from (this.message)
}

// set sender
pub fn (mut this Email) add_sender(value &AddressList) {
	value.iterate(fn [this](fullname string, emailaddress string, cset string) {
		C.g_mime_message_add_mailbox(this.message, C.GMimeAddressType(C.GMIME_ADDRESS_TYPE_SENDER), cstr(fullname), cstr(emailaddress))
	})
}

pub fn (mut this Email) get_sender() &C._InternetAddressList {
	return C.g_mime_message_get_sender (this.message)
}

// set reply to
pub fn (mut this Email) add_reply_to(value &AddressList) {
	value.iterate(fn [this](fullname string, emailaddress string, cset string) {
		C.g_mime_message_add_mailbox(this.message, C.GMimeAddressType(C.GMIME_ADDRESS_TYPE_REPLY_TO), cstr(fullname), cstr(emailaddress))
	})
}

pub fn (mut this Email) get_reply_to() &C._InternetAddressList {
	return C.g_mime_message_get_reply_to (this.message)
}

// set subject
pub fn (mut this Email) set_subject(subject string) {
	this.logger.info("set subject=$subject")
	sbjct := cstr(subject)
	C.g_mime_message_set_subject(this.message, sbjct, ccharset)
}

pub fn (mut this Email) get_subject() string {
	sbjct := C.g_mime_message_get_subject(this.message)
	if sbjct != voidptr(0) {
		subject := vstr(sbjct)
		this.logger.info("get subject=$subject")
		return subject
	}
	return ""
}

// set user agent
pub fn (mut this Email) set_user_agent(user_agent string) {
	C.g_mime_object_set_header(C.GMIME_OBJECT(this.message), cstr("User-Agent"), cstr(user_agent), ccharset)
}

// generate message id by suffix
pub fn (mut this Email) set_message_id(suffix string) {
	C.g_mime_message_set_message_id(this.message, C.g_mime_utils_generate_message_id(cstr(suffix)))
}

// set replied message id
pub fn (mut this Email) set_in_reply_to(origin_message_id string) {
	C.g_mime_object_set_header(C.GMIME_OBJECT(this.message), cstr("In-Reply-To"), cstr(origin_message_id), ccharset)
}

// set referenced message id
//pub fn (mut this Email) set_references(value string) {
//}
//
// set additional headers
pub fn (mut this Email) set_header(headername string, value string) {
	C.g_mime_object_set_header(C.GMIME_OBJECT(this.message), cstr(headername), cstr(value), ccharset)
}

// get additional headers
pub fn (mut this Email) get_header(headername string) string {
	headervalue := C.g_mime_object_get_header(C.GMIME_OBJECT(this.message), cstr(headername))
	if headervalue != voidptr(0) {
		return vstr(headervalue)
	}
	return ""
}

// set current date
pub fn (mut this Email) set_date_now() {
	date := C.g_date_time_new_from_unix_utc(int(C.time(/*C.NULL*/0)))
	C.g_mime_message_set_date(this.message, date)
	C.g_date_time_unref(date)
}

// set plain text
pub fn (mut this Email) set_text(text string, plain bool) {
	textpart := C.g_mime_text_part_new()
	defer { C.g_object_unref(C.G_OBJECT(textpart)) }
	C.g_mime_text_part_set_text(textpart, cstr(text))
	if plain {
		C.g_mime_text_part_set_charset(textpart, ccharset)
		C.g_mime_part_set_content_encoding(C.GMIME_PART(textpart), C.GMimeContentEncoding(C.GMIME_CONTENT_ENCODING_8BIT))
		C.g_mime_message_set_mime_part(this.message, C.GMIME_OBJECT(textpart))
	} else {
		C.g_mime_multipart_add(this.multipart, C.GMIME_OBJECT(textpart))
	}
}

pub fn (mut this Email) get_text() string {
	mut text := ""
	mut rtext := &text
	this.walk(fn [mut rtext, mut this](obj &C._GMimeObject) bool {
		//this.logger.debug("obj")
		if C.GMIME_IS_TEXT_PART(obj) != 0 {
			unsafe { *rtext = vstr(C.g_mime_text_part_get_text(C.GMIME_TEXT_PART(obj))) }
			this.logger.info("text '$rtext'")
			return false
		}
		return true
	})
	this.logger.debug("text '$text'")
	return text
}

// walk email parts
pub fn (mut this Email) walk(callback fn (&C._GMimeObject) bool) {
	iter := C.g_mime_part_iter_new (C.GMIME_OBJECT(this.message))
	defer { C.g_mime_part_iter_free (iter) }
	mut more := true
	for more {
		part := C.g_mime_part_iter_get_current (iter)
		more = callback(part)
		if more {
			more = C.g_mime_part_iter_next (iter) != 0
		}
	}
}

// encrypt email
pub fn (mut this Email) encrypt() bool {
	multipart := C.g_mime_multipart_new_with_subtype(cstr("mixed"))
	mut ret := false
	ctx := C.g_mime_gpg_context_new()
	defer { C.g_object_unref(C.G_OBJECT(ctx)) }
	// determine all recipients
	recipients := C.g_ptr_array_new()
	defer { C.g_ptr_array_free(recipients, /*C.TRUE*/1) }
	list := C.g_mime_message_get_all_recipients (this.message)
	defer { C.g_object_unref(C.G_OBJECT(list)) }
	count := C.internet_address_list_length (list)
	for i in 0 .. count {
		address := C.internet_address_list_get_address (list, i)
		C.g_ptr_array_add(recipients, C.internet_address_mailbox_get_addr(C.INTERNET_ADDRESS_MAILBOX(address)))
	}
	// TODO C.g_ptr_array_add(recipients, cstr(myself)) // always encrypt for myself
	// try to encrypt for all recipients
	err := &C._GError(0)
	encrypted := C.g_mime_multipart_encrypted_encrypt(ctx, C.G_OBJECT(multipart), /*FALSE*/0, voidptr(0), 0, recipients, &err)
	if encrypted == voidptr(0) {
		// encryption failed
		m := vstr(err.message)
		this.logger.info("encryption failed: '$m'")
		C.g_error_free(err)
		// plain
		C.g_mime_message_set_mime_part(this.message, C.GMIME_OBJECT(multipart))
	} else {
		// encrypted
		C.g_mime_message_set_mime_part(this.message, C.GMIME_OBJECT(encrypted))
		C.g_object_unref(C.G_OBJECT(encrypted))
		ret = true
	}
	return ret
}

// kick off editor
pub fn (mut this Email) edit() {
	// create temp file
	mut file, tempfile := util.temp_file(util.TempFileOptions{pattern: "epistula.vomposer."}) or { panic("temp_file failed") }
	file.close()
	err := &C._GError(0)
	stream := C.g_mime_stream_file_open(cstr(tempfile), cstr("w"), &err)
	if stream == voidptr(0) { panic(err.message) }
	defer { C.g_object_unref(C.G_OBJECT(stream)) }
	format := C.g_mime_format_options_get_default()
	C.g_mime_format_options_set_newline_format(format, C.GMimeNewLineFormat(C.GMIME_NEWLINE_FORMAT_DOS))
	written := C.g_mime_object_write_to_stream(C.GMIME_OBJECT(this.message), format, stream)
	if written <= 0 { panic('no bytes written') }
	this.logger.debug("written=$written")
	C.g_mime_stream_close(stream)
	// kick off editor
	editor := "/usr/bin/nvim"
	mut p := os.new_process(editor)
	p.set_args([
		"+set ft=mail", // switch to email syntax
		"+set fileencoding=" + charset, // use same encoding
		"+set enc=utf-8", // use utf8
		"+set fo+=w", // do wsf
		"+set fo-=ro", // dont repeat ">.." on new lines
		"+/^$", // jump to line after headers
		tempfile,
	])
	p.run()
	p.wait()
	p.close()
	this.parse(tempfile)
}

pub fn (mut this Email) attach(filename string) {
	err := &C._GError(0)
	// detect content time, the hard way
	mut type_ := "application"
	mut subtype := "octet-stream"
	file := C.g_file_new_for_path(cstr(filename))
	if file != voidptr(0) {
		defer { C.g_object_unref(C.G_OBJECT(file)) }
		file_info := C.g_file_query_info(file,
			cstr("standard::content-type,standard::type")/*C.G_FILE_ATTRIBUTE_STANDARD_TYPE "," C.G_FILE_ATTRIBUTE_STANDARD_CONTENT_TYPE*/,
			0/*G_FILE_QUERY_INFO_NONE*/, voidptr(0), &err)
		if file_info != voidptr(0) {
			defer { C.g_object_unref(C.G_OBJECT(file_info)) }
			ct := C.g_file_info_get_content_type(file_info)
			if ct != voidptr(0) {
				tp := C.g_mime_content_type_parse(C.g_mime_parser_options_get_default(), ct)
				mt := C.g_mime_content_type_get_media_type(tp)
				ms := C.g_mime_content_type_get_media_subtype(tp)
				if mt != voidptr(0) && ms != voidptr(0) {
					type_ = vstr(mt)
					subtype = vstr(ms)
				}
			}
		} else {
			this.logger.info("no file_info $err")
		}
	} else {
		this.logger.info("no file")
	}
	this.logger.info("content type '$type_/$subtype'")
	//
	part := C.g_mime_part_new_with_type(cstr(type_), cstr(subtype))
	defer { C.g_object_unref(C.G_OBJECT(part)) }
	C.g_mime_part_set_filename(part, cstr(os.base(filename)))
	// attach content
	stream := C.g_mime_stream_fs_open(cstr(filename), /*C.O_RDONLY*/0, 0644, &err)
	if stream == voidptr(0) {
		this.logger.info("file $filename not attached, $err.message")
		return
	}
	defer { C.g_object_unref(C.G_OBJECT(stream)) }
	content := C.g_mime_data_wrapper_new_with_stream(stream, C.GMimeContentEncoding(C.GMIME_CONTENT_ENCODING_DEFAULT))
	defer { C.g_object_unref(C.G_OBJECT(content)) }
	C.g_mime_part_set_content(part, content)
	C.g_mime_part_set_content_encoding(part, C.GMimeContentEncoding(C.GMIME_CONTENT_ENCODING_BASE64))
	// C.g_mime_part_set_content_description(part,
	// C.g_mime_part_set_content_id(part,
	// C.g_mime_part_set_content_md5(part,
	// C.g_mime_part_set_content_location(part,
	C.g_mime_multipart_add(this.multipart, C.GMIME_OBJECT(part))
	this.logger.info("terminate")
}

pub fn (mut this Email) transfer() int {
	format := C.g_mime_format_options_get_default()
	C.g_mime_format_options_set_newline_format(format, C.GMimeNewLineFormat(C.GMIME_NEWLINE_FORMAT_DOS))
	mailstring := C.g_mime_object_to_string(C.GMIME_OBJECT(this.message), format)
	if mailstring == voidptr(0) {
		this.logger.info("error getting mail as char buffer")
		return -1
	}
	buffer := vstr(mailstring)
	for commandline in [
		// ["/tmp/test.sh", "/tmp/test.eml", ], // test
		["/usr/sbin/sendmail", "-t", ], // transfer / send the email
		["/usr/bin/notmuch", "insert", "--decrypt=true", "+sent", "+inbox"], // store the email locally
	] {
		this.logger.info("running process $commandline")
		mut process := os.new_process(commandline[0])
		process.set_args(commandline[1..])
		process.set_redirect_stdio()
		process.run()
		process.stdin_write(buffer)
		os.fd_close(process.stdio_fd[0])
		o := process.stdout_slurp()
		if o.len > 0 { this.logger.info("stdout $o") }
		e := process.stderr_slurp()
		if e.len > 0 { this.logger.info("stderr $e") }
		process.wait()
		process.close()
		if process.code > 0 {
			err := process.err
			code := process.code
			this.logger.info("error running process $commandline: $code '$err'")
			return process.code
		}
		this.logger.info("done step")
	}
	this.logger.info("done transfering email")
	return 0
}

//	C.g_mime_multipart_add(multipart, C.GMIME_OBJECT(textpart))
//	mail_walk(mmsg, fn (part &C._GMimeObject) bool {
//		ct := C.g_mime_object_get_content_type (C.GMIME_OBJECT(part))
//		s := vstr(C.g_mime_content_type_get_mime_type (ct))
//		return true
//	})
