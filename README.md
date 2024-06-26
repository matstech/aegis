# Aegis

<p align="center">
  <img src="images/logo.jpeg" alt="Logo" width="320px" height="320px">
</p>

## Summary

## Description

*aegis* proposes a minimal authentication protocol between the client service and the server it is going to proxy. It works based on headers containing signatures.

## The protocol

Every time the client makes a request to the server service, it must be add the following headers:

- Auth-Headers: list of headers the client intends to sign. The list element seprator has to be *;*
For example: header<sub>1</sub>;....;header<sub>1</sub>
- Auth-Kid: the name of client service which has to be used from server to verify identity
- Auth-CorrelationId: it's the only *mandatory* header. It can be populated whit random value.
- Signature: the signature of the following string
  
      <list of indicated auth headers value>:Hash(payload)

> 👉 If the request does not expect a body or custom headers, payload or custom headers can be avoided. The signature must be composed by Auth-CorrelationId, at least.

## Algorithms

### Payload hash

As described in previous paragraph, payload has to be part of signature when exists. It is not intended as plain body but its computed hash.
Aegis expects to parse the [xxHash](https://xxhash.com/) of request body. In particular, the **XXH64** flavor is expected.

### Signature

Signature of described information has to be computed usign classical HMAC algorithm based on SHA512 hash function. A symmetric key was shared betwwen system (proxy and client) and this is used as HMAC key.

In the end, the signature header (in a complete scenario) will be as follows;

```
 HMAC-SHA512(Auth-CorrelationId;Auth-Headers...:XXH64(request body), symmetric key)
```

## Trust

There is no dynamic trust mode between client and server. Keys trusting will be statically a priori.

## Configuration

Configuration is based on [aconfig](https://github.com/cristalhq/aconfig) Go module. It has tobe in JSON form, example:

```json
{
    "ginmode": "debug", # release in production mode
    "loglevel": "debug",
    "server": {
        "mode": "PLAIN", # can be PLAIN,TLS,MTLS
        "tls": { # section needed just in TLS/MTLS case
            "certpath": "test/server.crt", # file path of server certificate
            "keypath": "test/server.key", # file path of key associated with server certificate
            "cacert": "test/cacert.pem" # file path of CAs for certificate verification (MTLS)
        },
        "port": 8080, # proxy listen port
        "probesport": 2112, # server port for probes endpoint (Kubernetes feat)
        "upstream": "httpbin.org" # upstream host
    },
    "kids": [ # list of strings representing all registered key id
        "test"
    ]
}
```

> 👉 Each kid MUST be associated to its value stored in specific env which has to be of the form *ACCESSKEY_<UpperCase(kid)>*. For example: if the kid is "test", the associated env will be *ACCESSKEY_TEST="abcde1242352525gsgs"*

## How to run

- Build docker image locally or use official [docker image](https://hub.docker.com/repository/docker/matteos93/aegis)
- Run docker image as follows:

```bash
docker run <imagename> -e CONFIG_PATH="<path to folder containing config.json>" -e ACCESSKEY_<KID1>="<secret value>" ... -e ACCESSKEY_<KIDn>="<secret value>" -p 8080:8080
```

## Example

This is an example of POST request to target server containing Authentication headers:

```bash
curl --location --request POST 'http://localhost:8080/post' \
--header 'Auth-CorrelationId: jxW7faeiNP' \
--header 'Auth-Kid: test' \
--header 'Signature: sM4EOA3jh/F7X3PqKI52Cr3Sa9kvS9YwkSSqFKGy3hExBrfPKoro3w3eJSq26Yw7I7ydesiXgcjxkMGLMVfiNQ==' \
--header 'Content-Type: application/json' \
--data-raw '{
    "min_position": 5,
    "has_more_items": false,
    "items_html": "Bus",
    "new_latent_count": 8,
    "data": {
        "length": 25,
        "text": "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat."
    },
    "numericalArray": [
        29,
        32,
        33,
        33,
        28
    ],
    "StringArray": [
        "Oxygen",
        "Oxygen",
        "Oxygen",
        "Oxygen"
    ],
    "multipleTypesArray": true,
    "objArray": [
        {
            "class": "lower",
            "age": 0
        },
        {
            "class": "upper",
            "age": 9
        },
        {
            "class": "middle",
            "age": 0
        },
        {
            "class": "lower",
            "age": 2
        },
        {
            "class": "lower",
            "age": 5
        }
    ]
}'
```

## Helm

Aegis has an official Helm Chart distribution which is available here and documented in this repository ([helm docs](aegis/README.md)).
