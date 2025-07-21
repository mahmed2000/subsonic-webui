# subsonic-webui

A feature-incomplete subsonic music "player"

It... exists, unfortunately.

# Usage

Its a single HTML page. Open it as a file, serve it over a bash script/python flask app/rust axum server/from nginx/idc. Just make sure CORS allows loading the following:

The only dependency is on https://pajhome.org.uk/crypt/md5/md5.js. If for some reason that goes down, hook up the script include to the copy in this repo and however you serve it.

Opens to an auth modal. Placeholder values will be set to the navidrome demo site as is. Creds also here for convenience:

Server: https://demo.navidrome.org
Username: demo
Password: demo

The server must support the token-salt auth method documented here: https://opensubsonic.netlify.app/docs/api-reference/. Basically means not [LMS](github.com/epoupon/lms).


