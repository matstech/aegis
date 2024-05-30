# go-token-guard

<p align="center">
  <img src="images/logo.png" alt="Logo" width="160px" height="160px">
</p>

## Summary

## Description

*go-token-guard* proposes a minimal authentication protocol between the client service and the server it is going to proxy. It works based on headers containing signatures.

## The protocol

Every time the client makes a request to the server service, it must be add the following headers:

- Auth-Headers: list of headers the client intends to sign. The list element seprator has to be *;*
For example: header<sub>1</sub>;....;header<sub>1</sub>
- Auth-Kid: the name of client service which has to be used from server to verify identity
- Auth-CorrelationId: it's the only mandatory header. It can be populated whit random value.
- Signature: the signature of the following string
  
      <list of indicated auth headers value>:Hash(payload)

> If the request does not expect a body, payload can be avoided and the signed string will be just the list of indicated auth headers value

## Algorithms

### Payload hash

### Signature

## Trust

## Configuration

## How to run

## Exit code

## Exit Code 1

- **Descrizione**: configurazioni sbagliate.
- **Causa**: Il programma è terminato correttamente senza errori.
- **Soluzione**: Nessuna azione necessaria.

## Exit Code 2

- **Descrizione**: errore nello start del server
- **Causa**: Il programma è terminato correttamente senza errori.
- **Soluzione**: Nessuna azione necessaria.
