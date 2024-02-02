# QUOTES TCP SERVER & CLIENT
## With Proof-of-Work verification

For the proof of work logic [Hashcash](https://en.wikipedia.org/wiki/Hashcash#:~:text=Hashcash%20is%20a%20cryptographic%20hash,proof%20can%20be%20verified%20efficiently.) algorithm is used (check the implementation inside `pkg/hashcash`).

To run the server, use command:
```
docker-compose up --build quotes-server
```

To run client:
```
docker-compose up --build quotes-client
```

Run unit tests:
```
go test -v ./...
```