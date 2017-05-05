Secrets bridge - Secure build-time secrets injection for Docker
===============================================================

Docker does not support build-time secrets, and this is a pain for any
`npm install`, gem installs, or whatever private repositories or
authenticated endpoints you need to contact during `docker build`
runs.

Also, you never want to have your credentials snapshotted in your
Docker image.

* It serves secrets defined on the host, either on the command-line or
  loaded from files.
* It acts as an SSH-Agent proxy, but secured through TLS, with
  temporary and auto-generated keypairs.

New version
-----------

secrets-bridge serve
# Serves the SSH Agent, with no secret keys. Useful on the otherwise with `secrets-bridge exec yarn`.
secrets-bridge serve --no-agent --secret key=value
# Prints out the bridge conf..
secrets-bridge exec -e THE_VALUE=key -- yarn
# Uses the bridge conf in `~/.bridge-conf` by default, can use the one via command-line with `--bridge-conf=BASE64STUFF`... Does the SSH agent forwarding automatically, on a random port anyway, won't do anything if the server doesn't serve it anyway.
# It fetches the `-e` and injects environment variables, instead of asking the client to fetch them.
secrets-bridge print key
# Reads the value stored in `bridge-conf`'s key and prints it out.
secrets-bridge test
# Tests the `bridge-conf` connection..


The _secrets bridge_ allows you to run a tiny server on your host as such:

    secrets-bridge-server serve -f ./bridge-conf \
                                -w \
                                --ssh-agent-forwarder \
                                --secret key=value \
                                --secret key2=value2 \
                                --secret-from-file key3=filename \
                                --timeout=300 &

and then, with a `Dockerfile` similar to this:

    RUN wget https://github.com/.../releases/.../secrets-bridge
    ARG BRIDGE_CONF
    RUN secrets-bridge -c ${BRIDGE_CONF} test
    RUN secrets-bridge -c ${BRIDGE_CONF} exec -- npm install
    RUN secrets-bridge -c ${BRIDGE_CONF} exec --no-agent -e SECRET=key -- ./do_sensitive_things.sh

run `docker build`:

    docker build --build-args BRIDGE_CONF=`cat bridge-conf` -t image/tag123 .

and, on the host, finish with:

    secrets-bridge kill -c $BRIDGE_CONF

to terminate the server.

## Other uses

With a bridge configuration (in base64), you can also:

    secrets-bridge serve -w -A

copy your `~/.bridge-conf` to the other location's `~/.bridge-conf` and then run:

    secrets-bridge exec ssh gcloud

over there.


### The `bridge-conf` file

The `bridge-conf` file contains a base64-encoded version of:

    {"endpoints": ["https://127.0.0.1:12345", "https://192.168.0.6:12345", "https://172.17.0.1:12345", "https://192.168.99.1:12345"],
     "cacert": "------ BEGIN CERTIFICATE -----\n...",
     "client_cert": "----- BEGIN CERTIFICATE -----\n...",
     "client_key": "----- BEGIN RSA PRIVATE KEY -----\n..."}

It allows the `secrets-bridge` inside the build-time container,
to communicate with the host, authenticate with the secrets server
and obtain credentials that were passed on the command line.

All of the information in this file is temporary and will vanish once
the server terminates. A self-signed CA and client cert/key pair is
generated on each `serve` runs.


### Installation

Download and install [https://golang.org/dl](Golang).  Install with:

```
go get github.com/abourget/secrets-bridge
```

This will build the `secrets-bridge` binary.  You will need a Linux
amd64 version for inside the containers. I'll soon release binaries in
the GitHub releases for quick download.


### Features

* SSH-Agent forwarding. The `client` sets the `SSH_AUTH_SOCK`
  environment variable when calling the sub-processes, and
  transparently passes that through the bridge, so the SSH-Agent on
  the host machine can serve the signing requests.

* Binary safe secrets

* Supports multi-line files

* On-the-fly base64 encoding and decoding of secrets, with `key` prefixes: `b64:` and `b64u:` for standard base64-encoding, and URL-safe encoding respectively.
  * Querying a prefixed key encodes its output
  * Setting a `key` (with `--secret b64:keyname=value`) will decode the passed-in `value` as base64, and store the original value, ready to be encoded upon query.

Encoding/decoding example:

```
secrets-bridge-server --secret b64:multilinekey=AAABCDEFG --secret theword=merde
```

Consume with:

```
curl http://localhost:9999/secrets/multilinekey  # non-base64 version of the secret, multiline
curl http://localhost:9999/secrets/b64u:multilinekey  # base64-url-safe version
curl http://localhost:9999/secrets/b64:multilinekey  # base64 standard version
curl http://localhost:9999/secrets/b64:theword  # base64-encoded "merde"
```


### Roadmap

* `--secret-to-file key=output_filename` to write files temporarily on
  `exec`, and clean-up after, before the next Docker layer snapshot.
* Implement `client -w CONF download file.key`..
* Implement `server --secret-from-file=[filename_only]`, reads the file and sets the key at the same time.
