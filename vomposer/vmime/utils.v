module vmime

const (
	charset = cstr("UTF-8")
)

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

