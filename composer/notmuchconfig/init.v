module notmuchconfig
import os

fn init() {
}

pub struct Config {
pub mut:
	database_path string
	user_name string
	user_primary_email string
}

pub fn new_config() Config {
	mut cfg := Config{}
	mut lines := os.read_lines(os.join_path(os.environ()['HOME'], ".notmuch-config")) or { panic(err) }
	mut section := ""
	for line in lines {
		if line.len > 0 && line[0] != `#` { // ignore empty and comment links
			if line[0] == `[` { // section delimiter
				section = line[1..line.len-1]
			} else {
				kv := line.split_nth("=", 2)
				key := kv[0]
				value := kv[1]
				if section == 'user' {
					if key == 'name' {
						cfg.user_name = value
					} else if key == 'primary_email' {
						cfg.user_primary_email = value
					}
				} else if section == 'database' {
					if key == 'path' {
						cfg.database_path = value
					}
				}
			}
		}
	}
	return cfg
}
