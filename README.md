# Peerster 


## Install

### Peerster

To install peerster, make sure that 'mux' and 'dedis/protobuf' are installed and present in your `$GOPATH`. Then, type `go build`. To build the client, `cd` into the folder `client` and execute `go build`.
One new command line option is available: `-mine-flood`. When activated, we will mine continously new blocks. Otherwise, we will only mine new blocks when they are non empty.

### Graphic Frontend

`Cd` inside `gui`.

**If you want to avoid installing npm, the frontend is already pre-compiled.**

Make sure you got `npm` installed. Then to install the dependencies execute `npm install`. To build the frontend then run `npm run build`.

To send a private message to an other peer, select the destination using the dropdown at the left of the title "Messages". A new tab should open where you will be able to send a message.

## How to use

Both the peerster and the client are used as presented in the homework.
To use the **graphical frontend** you need to launch the peerster with a `UIPort` of `8080`. The peerster will automatically launch the server, and you will be able to access to the frontend on`http://127.0.0.1:8080`.


## Code architecture

The code is divided in several parts:

- `gui/src`: code for the frontend using vue.js
- `main.go`: main entrypoint for peerster
- `client/`: client's code
- `lib/`: contains every abstraction used to implement peerster. In particular:
    - `gossiper.go`: the main logic of the protocol
    * `state.go`: a set of methods action on `State`, a structure representing the world known to a local peerster
    - `webserver.go`: the webserver for the frontend
    - `sparseSequence.go`: a datastructure to answer quickly to query of the type "which is the first non present element in a sequence". Discarded because I wrote it while non fully understanding the subject.
    - `file.go`: every functions having something to do with file upload and download (splitting in chunk, file reconstruction...)
    - `blockchain.go`: implementation of the blockchain
    - `fileKnowledge.go`: represent the knowledge we have about our world. Allow to answer questions as 'which peer have this metafile?', 'which peer have chunk xxx'.
- `searchRequest.go`: Deal with pending file searches, and allows to broadcast answers to the matchins searches.
    - `mergeResult.go`: able to merge the result of two search results, thus allowing us to detect matches.


## Implementation details

### Inference of missing blocks for the blockchain

The main difficulty to implement a valid conscencius algorithm for the blockchain is to understand that as we will receive some blocks later than expected (ex: receive A, then the father of A), we might have a disconnected blockchain. When we finally get a block that will join disconnected part, we want to make sure that we update the longest chain accordingly.

### Unified point to point routing

The routing for point to point messages is implemented using an interface. This means that the routing algorithm is implemented once and can be used to route both PrivateMessage, DataReply or DataRequest.

The same logic (creation of an interface) was used for broadcast message with a hop limit.

### The harddrive is a hashtable

The peerster itself holds no internal information about the file stores on its instance. Instead, it possess a folder, `_tmp_XXX` (where `XXX` is the server name) which contains a list of files. The name of each file is the hash of this file. Then, when this instance receives a chunk (or metafile) request, it tests if a corresponding file is present. This has several advantages:
- if we restart an instance of peerster with the same server name, the list of "registered" file will be the same
- chunks that are downloaded are immediately available for further sharing
- the implementation is easy and unified

Most of the cost of accessing the hard drive might be hidden by the cost of sending a message on the network.

O

Misc note: `test_generated.sh` and `test_generated2.sh` weren't wrote by me. If you lanch `test_generated2.sh`, it is normal that some tests are put as failed (and they should be about the blockchain). To test the blockchain I used the scripts as a way to launch a given topology easily, and then cross checked the logs to visualize any absurdity. 
