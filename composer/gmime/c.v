module gmime

#flag -lgmime-3.0 -lgio-2.0 -lgobject-2.0 -lglib-2.0 // LDFLAGS=`pkg-config --libs gmime-3.0`
#flag -D_LARGEFILE64_SOURCE -pthread -I/usr/include/gmime-3.0 -I/usr/include/libmount -I/usr/include/blkid -I/usr/include/glib-2.0 -I/usr/lib/x86_64-linux-gnu/glib-2.0/include // CFLAGS=`pkg-config --cflags gmime-3.0`

#include "glib.h"
#include "glib/gstdio.h"
#include "gio/gio.h"
#include "gmime/gmime.h"

[typedef] struct C.GError { domain int code int message &char }
[typedef] struct C.GByteArray { data &char len int}
[typedef] struct C.GObject { }
[typedef] struct C.GFileInfo { }
[typedef] struct C.GFile { }
[typedef] struct C.GDateTime { }
[typedef] struct C.GMimeContentType { }
[typedef] struct C.GMimeCryptoContext { }
[typedef] struct C.GCancellable { }
[typedef] struct C.GMimeDataWrapper { }
[typedef] struct C.GMimeFormatOptions { }
[typedef] struct C.GMimeMessage { }
[typedef] struct C.GMimeMultipart { }
[typedef] struct C.GMimeObject { }
[typedef] struct C.GMimeParser { }
[typedef] struct C.GMimeParserOptions { }
[typedef] struct C.GMimePart { }
[typedef] struct C.GMimePartIter { }
[typedef] struct C.GMimeStream { }
[typedef] struct C.GMimeTextPart { }
[typedef] struct C.GPtrArray { }
[typedef] struct C.InternetAddress { }
[typedef] struct C.InternetAddressGroup { }
[typedef] struct C.InternetAddressList { }
[typedef] struct C.InternetAddressMailbox { }
fn C.GMIME_IS_STREAM(voidptr) int
fn C.GMIME_OBJECT(voidptr) &C.GMimeObject
fn C.GMIME_STREAM(voidptr) &C.GMimeStream
fn C.GMIME_PART(voidptr) &C.GMimePart
fn C.GMIME_STREAM_MEM(voidptr) &C.GMimeStreamMem
fn C.G_OBJECT(voidptr) &C.GObject
fn C.INTERNET_ADDRESS_MAILBOX(voidptr) &C.InternetAddressMailbox
fn C.INTERNET_ADDRESS_GROUP(voidptr) &C.InternetAddressGroup
fn C.GMIME_IS_TEXT_PART(voidptr) int
fn C.GMIME_TEXT_PART(voidptr) &C.GMimeTextPart
fn C.g_date_time_new_from_unix_utc(int) &C.GDateTime
fn C.g_date_time_unref(&C.GDateTime)
fn C.g_error_free(&C.GError)
fn C.g_mime_charset_map_shutdown()
fn C.g_mime_content_type_get_mime_type(&C.GMimeContentType) &char
fn C.g_mime_data_wrapper_new_with_stream(&C.GMimeStream, C.GMimeContentEncoding) &C.GMimeDataWrapper
fn C.g_mime_format_options_get_default() &C.GMimeFormatOptions
fn C.g_mime_format_options_set_newline_format(&C.GMimeFormatOptions, C.GMimeNewLineFormat)
fn C.g_mime_gpg_context_new() &C.GMimeCryptoContext
fn C.g_mime_init()
fn C.g_mime_message_add_mailbox(&C.GMimeMessage, C.GMimeAddressType, &char, &char)
fn C.g_mime_message_new(C.gboolean) &C.GMimeMessage
fn C.g_mime_message_set_date(&C.GMimeMessage, &C.GDateTime)
fn C.g_mime_message_set_message_id(&C.GMimeMessage, &char)
fn C.g_mime_message_set_mime_part(&C.GMimeMessage, &C.GMimeObject)
fn C.g_mime_message_set_subject(&C.GMimeMessage, &char, &char)
fn C.g_mime_message_get_subject(&C.GMimeMessage) &char
fn C.g_mime_multipart_add(&C.GMimeMultipart, &C.GMimeObject)
fn C.g_mime_multipart_encrypted_encrypt(&C.GMimeCryptoContext, &C.GObject, int, voidptr, int, &C.GPtrArray, &&C.GError) &C.GMimeMultipartEncrypted
fn C.g_mime_multipart_new_with_subtype(&char) &C.GMimeMultipart
fn C.g_mime_object_get_content_type(&C.GMimeObject) &C.GMimeContentType
fn C.g_mime_object_get_header(&C.GMimeObject, &char) &char
fn C.g_mime_object_set_header(&C.GMimeObject, &char, &char, &char)
fn C.g_mime_object_write_to_stream(&C.GMimeObject, &C.GMimeFormatOptions, &C.GMimeStream) int
fn C.g_mime_object_to_string(&C.GMimeObject, &C.GMimeFormatOptions) &char
fn C.g_mime_parser_construct_message(&C.GMimeParser, &C.GMimeParserOptions) &C.GMimeMessage
fn C.g_mime_parser_new_with_stream(&C.GMimeStream) &C.GMimeParser
fn C.g_mime_parser_options_get_default() &C.GMimeParserOptions
fn C.g_mime_part_iter_free(&C.GMimePartIter)
fn C.g_mime_part_iter_get_current(&C.GMimePartIter) &C.GMimeObject
fn C.g_mime_part_iter_new(&C.GMimeObject) &C.GMimePartIter
fn C.g_mime_part_iter_next(&C.GMimePartIter) int //gboolean
fn C.g_mime_part_new_with_type(&char, &char) &C.GMimePart
fn C.g_mime_part_set_content(&C.GMimePart, &C.GMimeDataWrapper)
fn C.g_mime_part_set_content_encoding(&C.GMimePart, C.GMimeContentEncoding)
fn C.g_mime_part_set_filename(&C.GMimePart, &char)
fn C.g_mime_shutdown()
fn C.g_mime_stream_fs_open(&char, int, int, &&C.GError) &C.GMimeStream
fn C.g_mime_stream_mem_get_byte_array(&C.GMimeStreamMem) &C.GByteArray
fn C.g_mime_stream_mem_new() &C.GMimeStream
fn C.g_mime_text_part_new_with_subtype(&char) &C.GMimeTextPart
fn C.g_mime_text_part_new() &C.GMimeTextPart
fn C.g_mime_text_part_get_text(&C.GMimeTextPart) &char
fn C.g_mime_text_part_set_charset(&C.GMimeTextPart, &char)
fn C.g_mime_text_part_set_text(&C.GMimeTextPart, &char)
fn C.g_mime_utils_generate_message_id(&char) &char
fn C.g_object_unref(&C.GObject)
fn C.g_ptr_array_add(&C.GPtrArray, &char)
fn C.g_ptr_array_free(&C.GPtrArray, int) &char
fn C.g_ptr_array_new() &C.GPtrArray
fn C.internet_address_get_charset(&C.InternetAddress) &char
fn C.internet_address_get_name(&C.InternetAddress) &char
fn C.internet_address_list_get_address(&C.InternetAddressList, int) &C.InternetAddress
fn C.internet_address_list_length(&C.InternetAddressList) int
fn C.internet_address_list_parse(&C.GMimeParserOptions, &char) &C.InternetAddressList
fn C.internet_address_mailbox_get_addr(&C.InternetAddressMailbox) &char
fn C.internet_address_mailbox_new(&char, &char) &C.InternetAddress
fn C.internet_address_group_new(&char) &C.InternetAddress
fn C.internet_address_group_get_members(&C.InternetAddressGroup)&C.InternetAddressList
fn C.internet_address_group_set_members()
fn C.internet_address_group_add_member()
fn C.g_byte_array_free(&C.GByteArray, int) &char
fn C.time(voidptr) int
fn C.g_mime_message_get_all_recipients(&C.GMimeMessage) &C.InternetAddressList
fn C.g_mime_stream_file_open(&char, &char, &&C.GError) &C.GMimeStream
fn C.g_mime_stream_close(&C.GMimeStream)
fn C.g_file_new_for_path(&char) &C.GFile
fn C.g_file_query_info(&C.GFile, &char, int/*C.GFileQueryInfoFlags*/, &C.GCancellable, &&C.GError) &C.GFileInfo
fn C.g_file_info_get_content_type(&C.GFileInfo) &char
fn C.g_mime_content_type_get_media_type(&C.GMimeContentType) &char
fn C.g_mime_content_type_get_media_subtype(&C.GMimeContentType) &char
fn C.g_mime_content_type_parse(&C.GMimeParserOptions, &char) &C.GMimeContentType
fn C.internet_address_list_append(&C.InternetAddressList, &C.InternetAddressList)
fn C.internet_address_list_clear(&C.InternetAddressList)
fn C.internet_address_list_new() &C.InternetAddressList
fn C.g_mime_message_get_from(&C.GMimeMessage) &C.InternetAddressList
fn C.g_mime_message_get_to(&C.GMimeMessage) &C.InternetAddressList
fn C.g_mime_message_get_cc(&C.GMimeMessage) &C.InternetAddressList
fn C.g_mime_message_get_bcc(&C.GMimeMessage) &C.InternetAddressList
fn C.g_mime_message_get_sender(&C.GMimeMessage) &C.InternetAddressList
fn C.g_mime_message_get_reply_to(&C.GMimeMessage) &C.InternetAddressList

fn cstr(s string) &char {
	return &char(s.str)
}

fn vstr(s &char) string {
	return unsafe { cstring_to_vstring(s) }
}

