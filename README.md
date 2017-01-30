Secrets bridge - Secure build-time secrets injection for Docker
===============================================================

Docker does not support build-time secrets, and this is a pain for any
`npm install`, gem installs, or whatever private repositories or
authenticated endpoints you need to contact during `docker build`
runs.

Also, you never want to have your credentials snapshotted in your
Docker image.

_Docker secrets bridge_ allows you to run a tiny server on your host as such:

    secrets-bridge-server serve --bridge-conf=bridge-conf \
                                --ssh-agent-forwarder \
                                --secret key=value \
                                --secret key2=value2 \
                                --secret-from-file key3=filename \
                                --timeout=300 &

and then, with a Dockerfile similar to this:

    RUN wget https://some-location-with/secrets-bridge-server
    ARG BRIDGE_CONF
    RUN secrets-bridge-client --bridge-conf=${BRIDGE_CONF} test
    RUN secrets-bridge-client --bridge-conf=${BRIDGE_CONF} exec --ssh-agent -- npm install
    RUN secrets-bridge-client --bridge-conf=${BRIDGE_CONF} exec --listen=9999 -- ./do_sensitive_things.sh

run `docker build`:

    docker build --build-args BRIDGE_CONF=`cat bridge-conf` -t image/tag123 .

and finish with:

    secrets-bridge-server kill --bridge-conf=bridge-conf

to terminate the server.

An example of a `do_sensitive_things.sh` file could be:

    #!/bin/bash
    PASSWORD=$(curl http://localhost:9999/secrets/key)
    echo $PASSWORD | curl -u username https://secure.example.com/private-package.tgz -O /root/private-package.tgz

With the `--listen=9999` option, `secrets-bridge-client` will listen
on the loopback interface inside of the build-time container, and
serve the secrets, fetching them from the host transparently and
securely.


### The `bridge-conf` file

The `bridge-conf` file contains a base64-encoded version of:

    {"endpoints": ["https://127.0.0.1:12345", "https://192.168.0.6:12345", "https://172.17.0.1:12345", "https://192.168.99.1:12345"],
     "cacert": "------ BEGIN CERTIFICATE -----\n...",
     "client_cert": "----- BEGIN CERTIFICATE -----\n...",
     "client_key": "----- BEGIN PRIVATE KEY -----\n..."}

It allows the `secrets-bridge-client` inside the build-time container,
to communicate with the host, authenticate with the secrets server
and obtain credentials that were passed on the command line.

All of the information in this file is temporary and will vanish once
the server terminates. A new self-signed CA and client cert/key pair
will be generated on the next run.


### Features

* SSH-Agent forwarding. The `client` sets the `SSH_AUTH_SOCK`
  environment variable when calling the sub-processes, and
  transparently passes that through the bridge, so the SSH-Agent on
  the host machine can serve the signing requests.

* Text-based secrets, multi-line files. Secrets representation support
  on-the-fly base64-encoding (and URL-safe base64 encoding) with a
  `b64:` and `b64u:` prefix to keys.

  Tell the server the key is already base64 encoded:

    secrets-bridge-server --secret b64:multilinekey=AAABCDEFG --secret theword=merde

  and consume it with:

    curl http://localhost:9999/secrets/multilinekey  # non-base64 version of the secret, multiline
    curl http://localhost:9999/secrets/b64u:multilinekey  # base64-url-safe version
    curl http://localhost:9999/secrets/b64:multilinekey  # base64 standard version
    curl http://localhost:9999/secrets/b64:theword  # base64-encoded "merde"
