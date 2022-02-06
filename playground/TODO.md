TODOs
==

- retrieve public keys for a set of email addresses, if not available locally retrieve from key server, all via gpgme
- compose a (reply-)email properly with all necessary meta information, all via gmime
- encrypt an email for all recepients, all via gpgme
- use notmuch config in composer too, support mailto:
- content type & utf8
- decode quoted-printable
- check pure text parts for html artefacts `"&[^;]*;"`
- write own mails to maildir & kick off notmuch new
- keep cursor position on refresh; know index by item, not number
- show tags per thread, email
- the composer should be able to reply on multiple emails.
- the "replied" tag must be set somewhere on the originated mail
- use ansimage to display thumbnails of images
- add [tmux](https://tmux.github.io/) integration example

see also
--

https://th-h.de/net/usenet/faqs/headerfaq/
https://stackoverflow.com/questions/60380150/encrypting-headers-s-mime-message-rfc822
https://www.iana.org/assignments/message-headers/message-headers.xhtml
Autocrypt: addr=...; prefer-encrypt=mutual; keydata=
notmuch config set index.header.Autocrypt Autocrypt && notmuch reindex \*
notmuch set --database ;; to copy conf to db

