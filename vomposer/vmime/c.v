module vmime

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
[heap] struct C._GMimePartIter { }
[heap] struct C._GMimeContentType { }
[heap] struct C._GMimeParser { }
[heap] struct C._GDateTime { }
[heap] struct C._GMimeParserOptions { }
[heap] struct C._GMimeCryptoContext { }
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
fn C.g_mime_gpg_context_new() &C._GMimeCryptoContext
fn C.g_mime_multipart_encrypted_encrypt(&C._GMimeCryptoContext, &C.GObject, int, voidptr, int, &C._GPtrArray, &&C._GError) &C.GMimeMultipartEncrypted
fn C.g_error_free(&C._GError)
fn C.GMIME_IS_STREAM(voidptr) int
fn C.g_mime_parser_new_with_stream(&C._GMimeStream) &C._GMimeParser
fn C.g_mime_parser_construct_message(&C._GMimeParser, &C._GMimeParserOptions) &C._GMimeMessage
fn C.g_mime_parser_options_get_default() &C._GMimeParserOptions
fn C.g_mime_part_iter_new (&C._GMimeObject) &C._GMimePartIter
fn C.g_mime_part_iter_free (&C._GMimePartIter)
fn C.g_mime_part_iter_get_current (&C._GMimePartIter) &C._GMimeObject
fn C.g_mime_part_iter_next (&C._GMimePartIter) int //gboolean
fn C.g_mime_object_get_content_type (&C._GMimeObject) &C._GMimeContentType
fn C.g_mime_content_type_get_mime_type (&C._GMimeContentType) &char

