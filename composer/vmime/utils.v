module vmime

const (
	charset = "UTF-8"
	ccharset = cstr(charset)
)

pub struct AddressList {
mut:
	address_list &C._InternetAddressList
}

// create new from session
pub fn (this &Session) address_list_new() &AddressList {
	return &AddressList{
		C.internet_address_list_new()
	}
}

// free associated mem
pub fn (mut this AddressList) close() {
	C.g_object_unref(C.G_OBJECT(this.address_list))
}

pub fn (this &AddressList) add(value string) {
	list := C.internet_address_list_parse(C.g_mime_parser_options_get_default(), cstr(value))
	if list != voidptr(0) {
		defer { C.g_object_unref(C.G_OBJECT(list)) }
		C.internet_address_list_append(this.address_list, list)
	}
}

pub fn (this &AddressList) set(list &C._InternetAddressList) {
	C.internet_address_list_clear(this.address_list)
	C.internet_address_list_append(this.address_list, list)
}

pub fn (this &AddressList) len() int {
	return C.internet_address_list_length(this.address_list)
}

fn (this &AddressList) iterate(callback fn(string, string, string)) {
		count := C.internet_address_list_length(this.address_list)
		for i in 0 .. count {
			address := C.internet_address_list_get_address(this.address_list, i)
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

fn parse_address(value string, callback fn(string, string, string)) {
	list := C.internet_address_list_parse(C.g_mime_parser_options_get_default(), cstr(value))
	if list != voidptr(0) {
		defer { C.g_object_unref(C.G_OBJECT(list)) }
		count := C.internet_address_list_length(list)
		for i in 0 .. count {
			address := C.internet_address_list_get_address(list, i)
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

