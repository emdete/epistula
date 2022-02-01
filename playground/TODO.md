TODOs
==

- remove tag unread on leave
- pipe sendmail
- retrieve public keys for a set of email addresses, if not available locally retrieve from key server, all via gpgme
- compose a (reply-)email properly with all necessary meta information, all via gmime
- encrypt an email for all recepients, all via gpgme
- have another program for email composing which spawns an editor, check if tcell screen suspend-resume is working
- being started with mail target (like from a browser mailto: url)
- content type & utf8
- query input & unicode
- write own mails to maildir & kick off notmuch new
- keep cursor position on refresh; know index by item, not number

see also
--

https://th-h.de/net/usenet/faqs/headerfaq/
https://stackoverflow.com/questions/60380150/encrypting-headers-s-mime-message-rfc822
https://www.iana.org/assignments/message-headers/message-headers.xhtml
Autocrypt: addr=...; prefer-encrypt=mutual; keydata=
notmuch config set index.header.Autocrypt Autocrypt && notmuch reindex \*
notmuch set --database ;; to copy conf to db

