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

## Basic usage

Serves the SSH Agent securely over the bridge. Prints the bridge config to stdout, so you can bring it to the other node by **copy & pasting**:

    secrets-bridge serve -A

Execute `yarn` (or `npm install` or whatever) and leverage the SSH Agent from the remote host automatically:

    secrets-bridge exec -c [pasted-base64-configuration-from-host-a] yarn

Serve only the secret `key`, no SSH Agent forwarding:

    secrets-bridge serve --secret key=value

Serve `key1` taking its value from `filename1` and serve `filename2` as key `filename2`:

    secrets-bridge serve --secret-from-file key1=filename1 --secret-from-file filename2

Prints out secret `key`. This will use the default bridge configuration file at `~/.bridge-conf` (unless you specify an explicit config as b64 with `-c`):

    secrets-bridge print key

Execute `my-command.sh` with the env var `THE_VALUE` set to the value of the secret `key`:

    secrets-bridge exec -e THE_VALUE=key -- my-command.sh

This one prints the secret but encodes it to base64 first (see below for other variations):

    secrets-bridge print b64:key

You can also serve a secret that is already base64 encoded, as plain-text:

    secrets-bridge serve -w --secret b64:key=aGVsbG8td29ybGQK
    ...
    secrets-bridge print key
    hello-world

## Daemonization

You can start `serve` as a daemon with:

    secrets-bridge serve -d daemon.log -A -w -f bridge-conf

This will daemonize and log outputs to `daemon.log` (with `-d`), it
will enable SSH-Agent forwarding (`-A`), write (`-w`) the bridge
config to `bridge-conf` (with `-f`).

You can then kill that instance with:

    secrets-bridge kill -c $(cat bridge-conf)

Et hop!


## Usage with Docker

The _secrets bridge_ allows you to run a tiny server on your host as such:

    secrets-bridge serve -d daemon.log
                         -f ./bridge-conf -w \
                         --ssh-agent-forwarder \
                         --secret key=value \
                         --secret-from-file key2=filename \
                         --timeout=300

and then, with a `Dockerfile` similar to this:

    RUN wget https://github.com/.../releases/.../secrets-bridge
    ARG BRIDGE_CONF
    RUN secrets-bridge -c ${BRIDGE_CONF} test
    RUN secrets-bridge -c ${BRIDGE_CONF} exec -- npm install
    RUN secrets-bridge -c ${BRIDGE_CONF} exec -e SECRET=key -- ./do_sensitive_things.sh

run `docker build`:

    docker build --build-args BRIDGE_CONF=`cat bridge-conf` -t image/tag123 .

and, on the host, finish with:

    secrets-bridge kill -c `cat bridge-conf`

to terminate the server.

## Manual usage

With a bridge configuration (in base64), you can also:

    secrets-bridge serve -w -A

copy your `~/.bridge-conf` to the other location's `~/.bridge-conf` and then run:

    secrets-bridge exec ssh gcloud

over there.


## Base64 encoding

On-the-fly base64 encoding **and** decoding of secrets.

Prefix secrets with:

  * `b64:` for standard base64-
  * `b64u:` for URL-safe base64 codec.
  * `rb64:` for padding-less standard base64 codec.
  * `rb64u:` for padding-less URL-safe base64 codec.

Secrets are binary-safe and support multi-line files.


## SSH-Agent forwarding

The `client` sets the `SSH_AUTH_SOCK` environment variable when
calling the sub-processes, and transparently passes that through the
bridge, so the SSH-Agent on the host machine can serve the signing
requests.


## The `bridge-conf` file

The `bridge-conf` file contains a gzipped, base64-encoded version of:

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

You can elect to RE-USE a CA and set of keys in a subsequent run with
`--ca-key-store`. NOTE THAT this lessens the security, as it makes the
keys less "throw-away", making them more appealing to steal.


# Installation - from GitHub Releases

Grab a file here and `chmod +x` it if on Linux/Darwin:

https://github.com/abourget/secrets-bridge/releases

# Installation - from source

Download and install [https://golang.org/dl](Golang).  Install with:

```
go get github.com/abourget/secrets-bridge
```

This will build the `secrets-bridge` binary.  You will need a Linux
amd64 version for inside the containers.
