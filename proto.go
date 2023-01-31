package spider

type Proto interface {
	// Pack writes the Message into the connection.
	Pack(*message) error
	// Unpack reads bytes from the connection to the Message.
	Unpack(*[]byte) (*message, error)
}
