module vmime

import os
import io.util

const (
	charset = cstr("UTF-8")
)

fn cstr(s string) &char {
	return &char(s.str)
}

fn parse_address(value string, callback fn(string, string, string)) {
	list := C.internet_address_list_parse(C.g_mime_parser_options_get_default(), cstr(value))
	if list != voidptr(0) {
		defer { C.g_object_unref(C.G_OBJECT(list)) }
		count := C.internet_address_list_length (list)
		for i in 0 .. count {
			address := C.internet_address_list_get_address (list, i)
			if address != voidptr(0) {
				name := unsafe { C.internet_address_get_name(address).vstring() }
				a := C.internet_address_get_charset(address)
				mut cset := ""
				if a != voidptr(0) {
					cset = unsafe { a.vstring() }
				}
				addr := unsafe { C.internet_address_mailbox_get_addr(C.INTERNET_ADDRESS_MAILBOX(address)).vstring() }
				callback(name, addr, cset)
			}
		}
	}
}

pub fn samplerun() {
	emailaddress := cstr("mdt@emdete.de")
	fullname := cstr("M. Dietrich")
	//
	message := C.g_mime_message_new(C.gboolean(1))
	// meta / header
	C.g_mime_message_add_mailbox(message, C.GMimeAddressType(C.GMIME_ADDRESS_TYPE_FROM), fullname, emailaddress)
	C.g_mime_message_add_mailbox(message, C.GMimeAddressType(C.GMIME_ADDRESS_TYPE_SENDER), fullname, emailaddress)
	C.g_mime_message_add_mailbox(message, C.GMimeAddressType(C.GMIME_ADDRESS_TYPE_REPLY_TO), fullname, emailaddress)
	C.g_mime_message_add_mailbox(message, C.GMimeAddressType(C.GMIME_ADDRESS_TYPE_TO), fullname, emailaddress)
	C.g_mime_message_add_mailbox(message, C.GMimeAddressType(C.GMIME_ADDRESS_TYPE_CC), fullname, emailaddress)
	C.g_mime_message_add_mailbox(message, C.GMimeAddressType(C.GMIME_ADDRESS_TYPE_BCC), fullname, emailaddress)
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
		//mail_attach(multipart, "../screenshot.png")
	}
	// encrypt
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
	/*
	mmsg := email_parse(tempfile)
	status := unsafe { C.g_mime_object_get_header(C.GMIME_OBJECT(mmsg), cstr("X-Epistula-Status")).vstring() }
	eprintln("$status")
	eprintln("email_parse $mmsg")
	mail_walk(mmsg, fn (part &C._GMimeObject) bool {
		ct := C.g_mime_object_get_content_type (C.GMIME_OBJECT(part))
		s := unsafe { C.g_mime_content_type_get_mime_type (ct).vstring() }
		eprintln("$s")
		return true
		})
	*/
	// fini
	C.g_object_unref(C.G_OBJECT(message))
	return
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

