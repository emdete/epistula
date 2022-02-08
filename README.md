Epistula
==

This is a MUA, a mail user agent, a program to read and write your emails.

The original source is maintained at [codeberg](https://codeberg.org/mdt/epistula), please do PRs & issues here.

Matters can be discussed in a Matrix room #epistula:emdete.de and IRC on libera #epistula.

Whats not contained
--

You need a program to get the emails to your computer, various solutions are available for that task (See [Mail fetchers and synchronizers](https://notmuchmail.org/software/).

Epistula is based on [Notmuch](https://notmuchmail.org/) which organizes and finds emails.

While Epistula is console based it needs a way to start another terminal for composing emails. For now this is hardcoded to be a graphical terminal `gnome-terminal` which needs to be installed.

The mails will be written with you favourite editor which is retreive from the environment variable `EDITOR` which defaults to [nvim](http://neovim.org/) if not set.

HTML parts are dumped to pure text using [elinks](http://elinks.cz/) which should be installed as well.

A local MTA, a mail transfer agent is needed to actually send the email after composing (i suggest using [opensmtpd](https://www.opensmtpd.org/).

Whats contained
--

Epistula consists of two parts:

- The email browser
- The email composer

The browser shows your emails as threads and allows input of [search terms](https://notmuchmail.org/manpages/notmuch-search-terms-7/). The composer just kicks of the editor with a prepared email. It has no own UI so you have to put in fields in the header (To, CC, Bcc, Subject, ..) and the mail body. After that the mail is given over to the MTA.

The composer can be used to server mailto: urls from a browser.

Usage
--

The browser has a simple set of keys to be controlled. The UI has three areas: The query input on top, the resulting list of threads on the left, the list of mails in the selected thread on the right. Keyboard input is routed to each of the areas like

- Query edit:
	- All normal characters
	- Left, right, home and end
	- Tab resets the query
	- Enter execute query
- Thread list:
	- Up and down (next previous thread)
	- Control up and down (page up/down)
	- Control A Archive, untag inbox
	- Control S Tag as spam
- List of mails in the thread:
	- Page-up and page-down (page through the displayed mails)
	- Control page-up and page-down (next, previous part in the selected mail)
	- Control J and K (next, previous mail in the thread)
	- Control O (open part, show more lines)
	- Control R (reply email)
	- Control B Bounce (not done yet)
	- Control F Forward (not done yet)
- Global:
	- Control C (compose new email)
	- Control L Refresh screen
	- Mouse wheel for scrolling and left button for selecting are bound
	- Esc terminate Epistula

The mail list shows all parts of an email. The first text part is "opened" and the first 12 lines are shown. If there are more lines that is indicated by "+". Other parts can be opened (is supported) by clicking the triangle or by navigating to that part with Control-J/K.

Build
--

You need the go compiler, Debian based systems install it with:

```
# apt install golang-go
```

and compile the two components with

```
$ cd browser
$ GOOS=linux go build
$ cd ../composer
$ GOOS=linux go build
$ cd ..
```

Install
--

Instead of installing the components i just symlink the executables for now (which introduces some security risk):

```
# ln -s `pwd`/composer/epistula-composer /usr/local/bin
# ln -s `pwd`/browser/epistula-browser /usr/local/bin
# ln -s `pwd`/epistula.desktop /usr/local/share/applications
```

Warning
--

This program is in early state and contains rough edges. Display is not always scrolling where you expect and composing mail is for nerds. Navigating and scrolling is not finished yet.

The program uses panic() on errors which immedialty terminate the program. The programs log into /tmp/epistula-browser.log and /tmp/epistula-composer.log and the panic will give an stacktrace.

