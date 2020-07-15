package connmanager

// checkIncomingConnections makes sure there's no more then maxIncoming incoming connections
// if there are - it randomly disconnects enough to go below that number
func (c *ConnectionManager) checkIncomingConnections(connSet connectionSet) {
	if len(connSet) <= c.maxIncoming {
		return
	}

	// randomly disconnect nodes until the number of incoming connections is smaller the maxIncoming
	for address, connection := range connSet {
		err := connection.Disconnect()
		if err != nil {
			log.Errorf("Error disconnecting from %s: %+v", address, err)
		}

		connSet.remove(connection)
	}
}
