nmail: Send and Receive Mail via NNCP
=====

# Overview
`nmail` lets you send and receive mail via
[NNCP](http://www.nncpgo.org/). `nmail` assumes that all NNCP nodes
live in a single TLD `.nncp`. `nmail` works in conjunction with your
MTA to route mail messages meant for a `.nncp` address to a neighbor
node through `nncp-exec`. `nmail` also has a receiving mode which
munges headers in the incoming mail so that replies to the mail from
your MUA should be routed to an appropriate `.nncp` domain, which
can then be routed by `nmail` and `nncp-exec` to your destination.

# Building
Simply run:
```
make
```

and as long as you have a version of Make and Go compiler version 1.13
or higher, you should have an executable named `nmail` in the current
directory.

# Usage
## Send Mode
In send mode, `nmail` receives mail on standard input and then sends
the mail to a destination node. `nmail` is meant to be used in
conjunction with your MTA.

### CLI Usage
```
nmail <-handle [sendmail]> <send> <user@neighbor.nncp>
```

`<send>` and `<-handle foo>` is optional. This will take standard
input, parse it as an email, change the `To` header into an ID form to
prevent leaking any alias related metadata, and then invoke the handle
on the destination node. By default it invokes the `sendmail` handle,
but the handle can be set to anything if the neighbor node's handle
differs.

On the remote end, your MTA should be configured to accept this mail
and route it to the local user `user` as specified in the argument.

### Examples
```
nmail alice@aliceserver.nncp
```

This invokes `sendmail` with standard input on `aliceserver`.

```
nmail send alice@aliceserver.nncp
```

This does the same as above.


```
nmail -handle nmail alice@aliceserver.nncp
```

This sends standard input to the handle `nmail` on `aliceserver`.

## Receive Mode
In receive mode, `nmail` reads mail from `nncp` on standard input and
munges the From header to an address that can be replied to using
`nmail`, and then outputs the new mail to standard output. To have
`nmail` deliver mail after receipt, you can pipe standard output to
`sendmail` and have it sent to your local MTA.

### CLI Usage
```
nmail [receive|recv] <recipient>
```

### Examples
```
nmail receive
```

or

```
nmail recv
```

outputs the incoming mail to standard output.

```
nmail receive | sendmail "$1"
```

This receives mail from `nncp` and then pipes it to `sendmail` and
sends it to `$1`. An example of this usage is provided in the
`contrib/` folder and can be used to receive mail.

## Contrib Scripts
`contrib/` has two scripts, one which wraps `nmail` and can send mail,
(`nmsend.sh`) and one which wraps `nmail` and can receive mail
(`nmrecv.sh`). These scripts can be used to interact with your MTA.
