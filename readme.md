#Not ready for use without heavy modification

The application works by first creating a slice of pingers using the parameters and enpoints provided in the configuration file.

The pingers are started concurrently and are responsible for creating pings which are then passed to the connection manager via the pingOutbox channel.

The connection manager is responsible for creating connections to the endpoints and sending the pings to the endpoints. The connection manager also receives the responses from the endpoints and passes them to the response handler via the responseInbox channel.