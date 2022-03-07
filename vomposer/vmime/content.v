module vmime

import os

struct Content {
	multipart &C._GMimeMultipart
	recipients &C._GPtrArray
}

pub fn (mut this Content) close() {
	C.g_ptr_array_free(this.recipients, /*C.TRUE*/1)
	C.g_object_unref(C.G_OBJECT(this.multipart))
}

pub fn (mut this Content) attach(filename string) {
	err := &C._GError(0)
	stream := C.g_mime_stream_fs_open(cstr(filename), /*C.O_RDONLY*/0, 0644, &err)
	if stream == voidptr(0) {
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
	C.g_mime_multipart_add(this.multipart, C.GMIME_OBJECT(part))
}

pub fn (mut this Content) add_recipient(addr string) {
	C.g_ptr_array_add(this.recipients, cstr(addr))
}

