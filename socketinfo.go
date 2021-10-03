package main

// SocketInfo represents Linux-internal information about a socket,
// in a form ready to insert into the database.
type socketInfo struct {
	uid             string
	id              string
	iNode           uint32
	userID, groupID uint32
	state           string
}
