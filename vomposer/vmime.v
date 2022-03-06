// see https://github.com/vlang/v/blob/master/doc/docs.md
// see https://modules.vlang.io/
import os
import io.util

#flag -lgmime-3.0 -lgio-2.0 -lgobject-2.0 -lglib-2.0 // LDFLAGS=`pkg-config --libs gmime-3.0`
#flag -D_LARGEFILE64_SOURCE -pthread -I/usr/include/gmime-3.0 -I/usr/include/libmount -I/usr/include/blkid -I/usr/include/glib-2.0 -I/usr/lib/x86_64-linux-gnu/glib-2.0/include // CFLAGS=`pkg-config --cflags gmime-3.0`

#include "glib.h"
#include "glib/gstdio.h"

#include "gmime/gmime.h"

[heap] struct C._GError {
mut:
	domain int
	code int
	message &char
}
[heap] struct C.GObject { }
[heap] struct C._GMimeTextPart { }
[heap] struct C._GMimeParser { }
[heap] struct C._GDateTime { }
[heap] struct C._GMimeParserOptions { }
[heap] struct C.GMimeCryptoContext { }
[heap] struct C._GMimeObject { }
[heap] struct C._GMimeDataWrapper { }
[heap] struct C._GMimeMessage { }
[heap] struct C._GMimeStream { }
[heap] struct C._GMimeMultipart { }
[heap] struct C._GMimePart { }
[heap] struct C._GPtrArray { }
[heap] struct C._GByteArray {
mut:
	data &char
	len int
}
[heap] struct C._GMimeFormatOptions { }
fn C.GMIME_OBJECT(voidptr) &C._GMimeObject
fn C.GMIME_STREAM_MEM(voidptr) &C._GMimeStreamMem
fn C.GMIME_STREAM(voidptr) &C._GMimeStream
fn C.G_OBJECT(voidptr) &C.GObject
fn C.g_mime_charset_map_shutdown()
fn C.g_mime_format_options_get_default() &C._GMimeFormatOptions
fn C.g_mime_format_options_set_newline_format(&C._GMimeFormatOptions, C.GMimeNewLineFormat)
fn C.g_mime_init()
fn C.g_mime_message_new(C.gboolean) &C._GMimeMessage
fn C.g_mime_object_write_to_stream(&C._GMimeObject, &C._GMimeFormatOptions, &C._GMimeStream)
fn C.g_mime_shutdown()
fn C.g_mime_stream_mem_get_byte_array(&C._GMimeStreamMem) &C._GByteArray
fn C.g_mime_stream_mem_new() &C._GMimeStream
fn C.g_object_unref(&C.GObject)
fn C.g_mime_message_add_mailbox(&C._GMimeMessage, C.GMimeAddressType, &char, &char)
fn C.g_mime_object_set_header(&C._GMimeObject, &char, &char, &char)
fn C.g_mime_object_get_header(&C._GMimeObject, &char) &char
fn C.g_mime_message_set_subject(&C._GMimeMessage, &char, &char)
fn C.g_mime_message_set_message_id(&C._GMimeMessage, &char)
fn C.g_mime_utils_generate_message_id(&char) &char
fn C.g_date_time_new_from_unix_utc(int) &C._GDateTime
fn C.time(voidptr) int
fn C.g_mime_message_set_date(&C._GMimeMessage, &C._GDateTime)
fn C.g_date_time_unref(&C._GDateTime)
fn C.g_mime_multipart_new_with_subtype(&char) &C._GMimeMultipart
fn C.g_mime_message_set_mime_part(&C._GMimeMessage, &C._GMimeObject)
fn C.g_mime_text_part_new_with_subtype(&char) &C._GMimeTextPart
fn C.g_mime_multipart_add(&C._GMimeMultipart, &C._GMimeObject)
fn C.g_mime_text_part_set_charset(&C._GMimeTextPart, &char)
fn C.g_mime_text_part_set_text(&C._GMimeTextPart, &char)
fn C.g_mime_message_set_mime_part(&C._GMimeMessage, &C._GMimeObject)
fn C.g_mime_part_new_with_type(&char, &char) &C._GMimePart
fn C.g_mime_part_set_content_encoding(&C._GMimePart, C.GMimeContentEncoding)
fn C.g_mime_part_set_filename(&C._GMimePart, &char)
fn C.g_mime_stream_fs_open(&char, int, int, &&C._GError) &C._GMimeStream
fn C.g_mime_data_wrapper_new_with_stream(&C._GMimeStream, C.GMimeContentEncoding) &C._GMimeDataWrapper
fn C.g_mime_part_set_content(&C._GMimePart, &C._GMimeDataWrapper)
fn C.g_ptr_array_new() &C._GPtrArray
fn C.g_ptr_array_free(&C._GPtrArray, int)
fn C.g_ptr_array_add(&C._GPtrArray, &char)
fn C.g_mime_gpg_context_new() &C.GMimeCryptoContext
fn C.g_mime_multipart_encrypted_encrypt(&C.GMimeCryptoContext, &C.GObject, int, voidptr, int, &C._GPtrArray, &&C._GError) &C.GMimeMultipartEncrypted
fn C.g_error_free(&C._GError)
fn C.GMIME_IS_STREAM(voidptr) int
fn C.gmime_parse() 
fn C.g_mime_parser_new_with_stream (&C._GMimeStream) &C._GMimeParser
fn C.g_mime_parser_construct_message (&C._GMimeParser, &C._GMimeParserOptions) &C._GMimeMessage
fn C.g_mime_parser_options_get_default () &C._GMimeParserOptions

fn cstr(s string) &char {
	return &char(s.str)
}

fn main() {
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
	s := unsafe { (buffer.data).vstring_with_len(buffer.len) }
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

fn mail_parse(filename string) &C._GMimeMessage {
	err := &C._GError(0)
	stream := C.g_mime_stream_fs_open (cstr(filename), /*O_RDONLY*/0, 0644, &err)
	if stream == voidptr(0) {
		m := unsafe { err.message.vstring() }
		eprintln("encryption failed: '$m'")
		return voidptr(0)
	}
	parser := C.g_mime_parser_new_with_stream (stream)
	C.g_object_unref (C.G_OBJECT(stream))
	message := C.g_mime_parser_construct_message (parser, /*NULL*/voidptr(0))
	C.g_object_unref (C.G_OBJECT(parser))
	return message
}

