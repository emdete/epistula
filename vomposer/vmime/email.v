module vmime

import io.util
import os

// representing an email
pub struct Email {
mut:
	session &Session
	message &C._GMimeMessage // gmime3 mail message structure
	multipart &C._GMimeMultipart // multipart content
}

// create new from session
pub fn (this &Session) email_new() &Email {
	message := C.g_mime_message_new(C.gboolean(1))
	return &Email{
		this
		message
		C.g_mime_multipart_new_with_subtype(cstr("mixed"))
	}
}

// parse file from session
pub fn (this &Session) email_parse(filename string) &C._GMimeMessage {
	err := &C._GError(0)
	stream := C.g_mime_stream_fs_open(cstr(filename), /*O_RDONLY*/0, 0644, &err)
	if stream == voidptr(0) { panic( unsafe { err.message.vstring() } ) }
	parser := C.g_mime_parser_new_with_stream(stream)
	C.g_object_unref(C.G_OBJECT(stream))
	message := C.g_mime_parser_construct_message(parser, /*NULL*/voidptr(0))
	C.g_object_unref(C.G_OBJECT(parser))
	return message
}

// free mem
pub fn (mut this Email) close() {
	C.g_object_unref(C.G_OBJECT(this.message))
	C.g_object_unref(C.G_OBJECT(this.multipart))

}

// set to
pub fn (mut this Email) add_to(value string) {
	parse_address(value, fn [mut this](fullname string, emailaddress string, cset string) {
		C.g_mime_message_add_mailbox(this.message, C.GMimeAddressType(C.GMIME_ADDRESS_TYPE_TO), cstr(fullname), cstr(emailaddress))
	})
}

// set carbon copy
pub fn (mut this Email) add_cc(value string) {
	parse_address(value, fn [mut this](fullname string, emailaddress string, cset string) {
		C.g_mime_message_add_mailbox(this.message, C.GMimeAddressType(C.GMIME_ADDRESS_TYPE_CC), cstr(fullname), cstr(emailaddress))
	})
}

// set blind carbon copy
pub fn (mut this Email) add_bcc(value string) {
	parse_address(value, fn [mut this](fullname string, emailaddress string, cset string) {
		C.g_mime_message_add_mailbox(this.message, C.GMimeAddressType(C.GMIME_ADDRESS_TYPE_BCC), cstr(fullname), cstr(emailaddress))
	})
}

// set from
pub fn (mut this Email) add_from(value string) {
	parse_address(value, fn [this](fullname string, emailaddress string, cset string) {
		C.g_mime_message_add_mailbox(this.message, C.GMimeAddressType(C.GMIME_ADDRESS_TYPE_FROM), cstr(fullname), cstr(emailaddress))
	})
}

// set sender
pub fn (mut this Email) add_sender(value string) {
	parse_address(value, fn [this](fullname string, emailaddress string, cset string) {
		C.g_mime_message_add_mailbox(this.message, C.GMimeAddressType(C.GMIME_ADDRESS_TYPE_SENDER), cstr(fullname), cstr(emailaddress))
	})
}

// set reply to
pub fn (mut this Email) add_reply_to(value string) {
	parse_address(value, fn [this](fullname string, emailaddress string, cset string) {
		C.g_mime_message_add_mailbox(this.message, C.GMimeAddressType(C.GMIME_ADDRESS_TYPE_REPLY_TO), cstr(fullname), cstr(emailaddress))
	})
}

// set subject
pub fn (mut this Email) set_subject(subject string) {
	C.g_mime_message_set_subject(this.message, cstr(subject), charset)
}

// set user agent
pub fn (mut this Email) set_user_agent(user_agent string) {
	C.g_mime_object_set_header(C.GMIME_OBJECT(this.message), cstr("User-Agent"), cstr(user_agent), charset)
}

// generate message id by suffix
pub fn (mut this Email) set_message_id(suffix string) {
	C.g_mime_message_set_message_id(this.message, C.g_mime_utils_generate_message_id(cstr(suffix)))
}

// set replied message id
pub fn (mut this Email) set_in_reply_to(origin_message_id string) {
	C.g_mime_object_set_header(C.GMIME_OBJECT(this.message), cstr("In-Reply-To"), cstr(origin_message_id), charset)
}

// set referenced message id
pub fn (mut this Email) set_references(value string) {
}

// set additional headers
pub fn (mut this Email) set_header_x(headername string, value string) {
	C.g_mime_object_set_header(C.GMIME_OBJECT(this.message), cstr(headername), cstr(value), charset)
}

// get additional headers
pub fn (mut this Email) get_header(headername string) string {
	headervalue := C.g_mime_object_get_header(C.GMIME_OBJECT(this.message), cstr(headername))
	if headervalue != voidptr(0) {
		return unsafe { headervalue.vstring() }
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
pub fn (mut this Email) set_text(text string) {
	textpart := C.g_mime_text_part_new_with_subtype(cstr("plain"))
	defer { C.g_object_unref(C.G_OBJECT(textpart)) }
	C.g_mime_text_part_set_text(textpart, cstr(text))
	C.g_mime_text_part_set_charset(textpart, charset)
	C.g_mime_part_set_content_encoding(C.GMIME_PART(textpart), C.GMimeContentEncoding(C.GMIME_CONTENT_ENCODING_8BIT))
	C.g_mime_message_set_mime_part(this.message, C.GMIME_OBJECT(textpart))
}

// walk email parts
pub fn (mut this Email) mail_walk(callback fn (&C._GMimeObject) bool) {
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
	mut ret := false
	ctx := C.g_mime_gpg_context_new()
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
	encrypted := C.g_mime_multipart_encrypted_encrypt(ctx, C.G_OBJECT(this.multipart), /*FALSE*/0, voidptr(0), 0, recipients, &err)
	if encrypted == voidptr(0) {
		// encryption failed
		m := unsafe { err.message.vstring() }
		eprintln("encryption failed: '$m'")
		C.g_error_free(err)
		// plain
		C.g_mime_message_set_mime_part(this.message, C.GMIME_OBJECT(this.multipart))
	} else {
		// encrypted
		C.g_mime_message_set_mime_part(this.message, C.GMIME_OBJECT(encrypted))
		C.g_object_unref(C.G_OBJECT(encrypted))
		ret = true
	}
	C.g_object_unref(C.G_OBJECT(ctx))
	return ret
}

// kick off editor
pub fn (mut this Email) edit() {
	// prepare mail
	this.set_header_x("MIME-Version", "1.0")
	//this.set_header_x("Content-Type", "text/plain; charset=utf-8")
	//this.set_header_x("Content-Transfer-Encoding", "8bit")
	mut filename := ''
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
	eprintln("written is $written")
	filename = tempfile
	C.g_mime_stream_close(stream)
	// kick off editor
	editor := "/usr/bin/nvim"
	mut p := os.new_process(editor)
	p.set_args([
		"+set ft=mail", // switch to email syntax
		"+set fileencoding=utf-8", // use utf8
		"+set enc=utf-8", // use utf8
		"+set fo+=w", // do wsf
		"+set fo-=ro", // dont repeat ">.." on new lines
		"+/^$", // jump to line after headers
		filename,
	])
	p.run()
	p.wait()
	p.close()
	C.g_object_unref(C.G_OBJECT(this.message))
	this.message = this.session.email_parse(filename)
}

pub fn (mut this Email) attach(filename string) {
	err := &C._GError(0)
	stream := C.g_mime_stream_fs_open(cstr(filename), /*C.O_RDONLY*/0, 0644, &err)
	if stream == voidptr(0) {
		return //error("file $filename not attached, $err.message")
	}
	defer { C.g_object_unref(C.G_OBJECT(stream)) }
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
					type_ = unsafe { mt.vstring() }
					subtype = unsafe { ms.vstring() }
				}
			}
		} else {
			eprintln("no file_info $err")
		}
	} else {
		eprintln("no file")
	}
	eprintln("content type '$type_/$subtype'")
	part := C.g_mime_part_new_with_type(cstr(type_), cstr(subtype))
	defer { C.g_object_unref(C.G_OBJECT(part)) }
	C.g_mime_part_set_filename(part, cstr(os.base(filename)))
	content := C.g_mime_data_wrapper_new_with_stream(stream, C.GMimeContentEncoding(C.GMIME_CONTENT_ENCODING_DEFAULT))
	defer { C.g_object_unref(C.G_OBJECT(content)) }
	C.g_mime_part_set_content(part, content)
	C.g_mime_part_set_content_encoding(part, C.GMimeContentEncoding(C.GMIME_CONTENT_ENCODING_BASE64))
	// C.g_mime_part_set_content_description(part,
	// C.g_mime_part_set_content_id(part,
	// C.g_mime_part_set_content_md5(part,
	// C.g_mime_part_set_content_location(part,
	C.g_mime_multipart_add(this.multipart, C.GMIME_OBJECT(part))
	//C.g_mime_message_set_mime_part(this.message, C.GMIME_OBJECT(this.multipart))
}

//	C.g_mime_multipart_add(multipart, C.GMIME_OBJECT(textpart))
//	mail_walk(mmsg, fn (part &C._GMimeObject) bool {
//		ct := C.g_mime_object_get_content_type (C.GMIME_OBJECT(part))
//		s := unsafe { C.g_mime_content_type_get_mime_type (ct).vstring() }
//		return true
//	})
