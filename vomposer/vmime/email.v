module vmime

// representing an email
pub struct Email {
mut:
	message &C._GMimeMessage
	content Content
}

// create new from session
pub fn (this &Session) email_new() &Email {
	message := C.g_mime_message_new(C.gboolean(1))
	C.g_mime_object_set_header(C.GMIME_OBJECT(this.message), cstr("User-Agent"), cstr(value), charset)
	return &Email{
		message
		Content {
			C.g_mime_multipart_new_with_subtype(cstr("mixed"))
			C.g_ptr_array_new()
		}
	}
}

// parse file from session
pub fn (this &Session) email_parse(filename string) &C._GMimeMessage {
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


pub fn (mut this Email) add_bcc(value string) {
	parse_address(value, fn [mut this](fullname string, emailaddress string, cset string) {
		C.g_mime_message_add_mailbox(this.message, C.GMimeAddressType(C.GMIME_ADDRESS_TYPE_BCC), cstr(fullname), cstr(emailaddress))
		this.content.add_recipient(emailaddress)
	})
}

pub fn (mut this Email) add_cc(value string) {
	parse_address(value, fn [mut this](fullname string, emailaddress string, cset string) {
		C.g_mime_message_add_mailbox(this.message, C.GMimeAddressType(C.GMIME_ADDRESS_TYPE_CC), cstr(fullname), cstr(emailaddress))
		this.content.add_recipient(emailaddress)
	})
}

pub fn (mut this Email) add_from(value string) {
	parse_address(value, fn [this](fullname string, emailaddress string, cset string) {
		C.g_mime_message_add_mailbox(this.message, C.GMimeAddressType(C.GMIME_ADDRESS_TYPE_FROM), cstr(fullname), cstr(emailaddress))
	})
}

pub fn (mut this Email) set_message_id(value string) {
	C.g_mime_message_set_message_id(message, C.g_mime_utils_generate_message_id(cstr(value)))
}

pub fn (mut this Email) set_references(value string) {
}

pub fn (mut this Email) add_sender(value string) {
	parse_address(value, fn [this](fullname string, emailaddress string, cset string) {
		C.g_mime_message_add_mailbox(this.message, C.GMimeAddressType(C.GMIME_ADDRESS_TYPE_SENDER), cstr(fullname), cstr(emailaddress))
	})
}

pub fn (mut this Email) add_reply_to(value string) {
	parse_address(value, fn [this](fullname string, emailaddress string, cset string) {
		C.g_mime_message_add_mailbox(this.message, C.GMimeAddressType(C.GMIME_ADDRESS_TYPE_REPLY_TO), cstr(fullname), cstr(emailaddress))
	})
}

pub fn (mut this Email) set_in_reply_to(id string) {
}

pub fn (mut this Email) set_subject(subject string) {
	C.g_mime_message_set_subject(this.message, cstr(subject), charset)
}

pub fn (mut this Email) add_to(value string) {
	parse_address(value, fn [mut this](fullname string, emailaddress string, cset string) {
		C.g_mime_message_add_mailbox(this.message, C.GMimeAddressType(C.GMIME_ADDRESS_TYPE_TO), cstr(fullname), cstr(emailaddress))
		this.content.add_recipient(emailaddress)
	})
}

pub fn (mut this Email) set_text_from_file(value string) {
}

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

pub fn (mut this Email) encrypt() {
	ctx := C.g_mime_gpg_context_new()
	err := &C._GError(0)
	encrypted := C.g_mime_multipart_encrypted_encrypt(ctx, C.G_OBJECT(this.content.multipart), /*FALSE*/0, voidptr(0), 0, this.content.recipients, &err)
	if encrypted == voidptr(0) {
		m := unsafe { err.message.vstring() }
		eprintln("encryption failed: '$m'")
		C.g_error_free(err)
		// plain
		C.g_mime_message_set_mime_part(this.message, C.GMIME_OBJECT(this.content.multipart))
	} else {
		// encrypted
		C.g_mime_message_set_mime_part(this.message, C.GMIME_OBJECT(encrypted))
		C.g_object_unref(C.G_OBJECT(encrypted))
	}
	C.g_object_unref(C.G_OBJECT(ctx))
}

