## Install

### Peerster

To install peerster, make sure that 'mux' and 'dedis/protobuf' are installed and present in your `$GOPATH`. Then, type `go build`. To build the client, `cd` into the folder `client` and execute `go build`.

### Graphic Frontend

`Cd` inside `gui`.

**If you want to avoid installing npm, the frontend is already pre-compiled.**

Make sure you got `npm` installed. Then to install the dependencies execute `npm install`. To build the frontend then run `npm run build`.

## How to use

Both the peerster and the client are used as presented in the homework.
To use the **graphical frontend** you need to launch the peerster with a `uiport` of `8080`. The peerster will automatically launch the server, and you will be able to access to the frontend on`http://127.0.0.1:8080`.
