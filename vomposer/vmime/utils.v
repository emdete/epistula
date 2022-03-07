module vmime

import os
import io.util

pub struct Email {
}

pub fn (mut this Email) add_bcc(value string) {
}

pub fn (mut this Email) add_cc(value string) {
}

pub fn (mut this Email) add_from(value string) {
}

pub fn (mut this Email) add_message_id(value string) {
}

pub fn (mut this Email) add_references(value string) {
}

pub fn (mut this Email) add_reply_to(value string) {
}

pub fn (mut this Email) add_in_reply_to(value string) {
}

pub fn (mut this Email) set_subject(value string) {
}

pub fn (mut this Email) add_to(value string) {
}

pub fn (mut this Email) set_text_from_file(value string) {
}

fn cstr(s string) &char {
	return &char(s.str)
}

pub fn samplerun() {
	myself := cstr("mdt@emdete.de")
	myname := cstr("M. Dietrich")
	charset := cstr("UTF-8")
	//
	C.g_mime_init()
	//
	message := C.g_mime_message_new(C.gboolean(1))
	// meta / header
	C.g_mime_message_add_mailbox(message, C.GMimeAddressType(C.GMIME_ADDRESS_TYPE_FROM), myname, myself)
	C.g_mime_message_add_mailbox(message, C.GMimeAddressType(C.GMIME_ADDRESS_TYPE_SENDER), myname, myself)
	C.g_mime_message_add_mailbox(message, C.GMimeAddressType(C.GMIME_ADDRESS_TYPE_REPLY_TO), myname, myself)
	C.g_mime_message_add_mailbox(message, C.GMimeAddressType(C.GMIME_ADDRESS_TYPE_TO), myname, myself)
	C.g_mime_message_add_mailbox(message, C.GMimeAddressType(C.GMIME_ADDRESS_TYPE_CC), myname, myself)
	C.g_mime_message_add_mailbox(message, C.GMimeAddressType(C.GMIME_ADDRESS_TYPE_BCC), myname, myself)
	C.g_mime_object_set_header(C.GMIME_OBJECT(message), cstr("User-Agent"), cstr("Epistula"), charset)
	C.g_mime_object_set_header(C.GMIME_OBJECT(message), cstr("X-Epistula-Status"), cstr("I am not done"), charset)
	date := C.g_date_time_new_from_unix_utc(int(C.time(/*C.NULL*/0)))
	C.g_mime_message_set_date(message, date)
	C.g_date_time_unref(date)
	C.g_mime_message_set_message_id(message, C.g_mime_utils_generate_message_id(cstr("epistula.de")))
	C.g_mime_message_set_subject(message, cstr("How you doin?"), charset)
	// body
	multipart := C.g_mime_multipart_new_with_subtype(cstr("mixed"))
	// textpart
	{
		textpart := C.g_mime_text_part_new_with_subtype(cstr("plain"))
		C.g_mime_text_part_set_charset(textpart, cstr("utf-8"))
		C.g_mime_text_part_set_text(textpart, cstr('Hey Alice,

What are you up to this weekend? Monica is throwing one of her parties on
Saturday and I was hoping you could make it.
Will you be my +1?

-- Joy
'))
		C.g_mime_multipart_add(multipart, C.GMIME_OBJECT(textpart))
		//C.g_mime_message_set_mime_part(message, C.GMIME_OBJECT(textpart)) //
		C.g_object_unref(C.G_OBJECT(textpart))
	}
	// attach
	{
		mail_attach(multipart, "../screenshot.png")
	}
	// encrypt
	{
		recipients := C.g_ptr_array_new()
		C.g_ptr_array_add(recipients, myself)
		C.g_ptr_array_add(recipients, cstr("test@sample.org"))
		ctx := C.g_mime_gpg_context_new()
		err := &C._GError(0)
		encrypted := C.g_mime_multipart_encrypted_encrypt(ctx, C.G_OBJECT(multipart), /*FALSE*/0, voidptr(0), 0, recipients, &err)
		if encrypted == voidptr(0) {
			m := unsafe { err.message.vstring() }
			eprintln("encryption failed: '$m'")
			C.g_error_free(err)
			// plain
			C.g_mime_message_set_mime_part(message, C.GMIME_OBJECT(multipart))
		} else {
			// encrypted
			C.g_mime_message_set_mime_part(message, C.GMIME_OBJECT(encrypted))
			C.g_object_unref(C.G_OBJECT(encrypted))
		}
		C.g_ptr_array_free(recipients, /*C.TRUE*/1)
		C.g_object_unref(C.G_OBJECT(ctx))
	}
	C.g_object_unref(C.G_OBJECT(multipart))
	// dump
	stream := C.g_mime_stream_mem_new()
	format := C.g_mime_format_options_get_default()
	C.g_mime_format_options_set_newline_format(format, C.GMimeNewLineFormat(C.GMIME_NEWLINE_FORMAT_DOS))
	C.g_mime_object_write_to_stream(C.GMIME_OBJECT(message), format, stream)
	buffer := C.g_mime_stream_mem_get_byte_array(C.GMIME_STREAM_MEM(stream))
	s := unsafe {
		(buffer.data).vstring_with_len(buffer.len) }
	C.g_object_unref(C.G_OBJECT(stream))
	mut file, tempfile := util.temp_file(util.TempFileOptions{pattern: "epistula.vomposer."}) or {
		eprintln("temp_file failed")
		exit(-1) }
	eprintln("$tempfile")
	file.write_string(s) or {
		eprintln("write_string failed")
		exit(-1) }
	file.close()
	mail_edit(tempfile)
	mmsg := mail_parse(tempfile)
	status := unsafe { C.g_mime_object_get_header(C.GMIME_OBJECT(mmsg), cstr("X-Epistula-Status")).vstring() }
	eprintln("$status")
	eprintln("mail_parse $mmsg")
	mail_walk(mmsg, fn (part &C._GMimeObject) bool {
		ct := C.g_mime_object_get_content_type (C.GMIME_OBJECT(part))
		s := unsafe { C.g_mime_content_type_get_mime_type (ct).vstring() }
		eprintln("$s")
		return true
		})
	// fini
	C.g_object_unref(C.G_OBJECT(message))
	C.g_mime_charset_map_shutdown()
	C.g_mime_shutdown()
	return
}

fn mail_attach(multipart &C._GMimeMultipart, filename string) { //?&C._GMimeMultipart {
	err := &C._GError(0)
	stream := C.g_mime_stream_fs_open(cstr(filename), /*C.O_RDONLY*/0, 0644, &err)
	if stream == 0 {
		return //error("file $filename not attached, $err.message")
	}
	defer { C.g_object_unref(C.G_OBJECT(stream)) }
	part := C.g_mime_part_new_with_type(cstr("image"), cstr("png"))
	defer { C.g_object_unref(C.G_OBJECT(part)) }
	C.g_mime_part_set_filename(part, cstr(os.base(filename)))
	content := C.g_mime_data_wrapper_new_with_stream(stream, C.GMimeContentEncoding(C.GMIME_CONTENT_ENCODING_DEFAULT))
	defer { C.g_object_unref(C.G_OBJECT(content)) }
	C.g_mime_part_set_content(part, content)
	C.g_mime_part_set_content_encoding(part, C.GMimeContentEncoding(C.GMIME_CONTENT_ENCODING_BASE64))
	C.g_mime_multipart_add(multipart, C.GMIME_OBJECT(part))
	//return multipart
}

fn mail_edit(filename string) {
	//
	editor := "/usr/bin/nvim"
	mut p := os.new_process(editor)
	p.set_args([
		"+set ft=mail", // switch to email syntax
		"+set fileencoding=utf-8", // use utf8
		"+set enc=utf-8", // use utf8
		"+set fo+=w", // do wsf
		"+set fo-=ro", // dont repeat ">.." on new lines
		filename,
	])
	p.run()
	p.wait()
	p.close()
}

fn mail_walk(message &C._GMimeMessage, callback fn (&C._GMimeObject) bool) {
	iter := C.g_mime_part_iter_new (C.GMIME_OBJECT(message))
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

fn mail_parse(filename string) &C._GMimeMessage {
	err := &C._GError(0)
	stream := C.g_mime_stream_fs_open(cstr(filename), /*O_RDONLY*/0, 0644, &err)
	if stream == voidptr(0) {
		m := unsafe { err.message.vstring() }
		eprintln("encryption failed: '$m'")
		return voidptr(0)
	}
	parser := C.g_mime_parser_new_with_stream(stream)
	C.g_object_unref(C.G_OBJECT(stream))
	message := C.g_mime_parser_construct_message(parser, /*NULL*/voidptr(0))
	C.g_object_unref(C.G_OBJECT(parser))
	return message
}


