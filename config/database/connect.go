package database

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/gocql/gocql"
)

// Connect establishes a connection to the Cassandra database.
// It creates a session and sets the keyspace to "urllite_dev".
// If the connection fails, it logs the error and exits the program.
// The session is closed after use to free up resources.

func Connect() {
	cluster := gocql.NewCluster(os.Getenv("CASSANDRA_HOST"))
	cassandraPort, err := strconv.Atoi(os.Getenv("CASSANDRA_PORT"))
	if err != nil {
		log.Fatal("Invalid Cassandra port:", err.Error())
	}
	cluster.Port = cassandraPort
	cluster.Keyspace = os.Getenv("CASSANDRA_SYSTEM_KEYSPACE")
	cluster.Consistency = gocql.Quorum
	session, err := cluster.CreateSession()
	if err != nil {
		log.Fatal("Unable to connect to Cassandra:", err.Error())
	}
	defer session.Close()

	// Set the keyspace to "urllite_dev"
	keyspace := os.Getenv("CASSANDRA_URLLITE_KEYSPACE")
	// Create the keyspace if it doesn't exist
	createKeyspaceQry := fmt.Sprintf("CREATE KEYSPACE IF NOT EXISTS %s WITH REPLICATION = { 'class' : 'SimpleStrategy', 'replication_factor' : 1 }", keyspace)
	err = session.Query(createKeyspaceQry).Exec()
	if err != nil {
		log.Fatal("Unable to create keyspace:", err.Error())
	}
	// Close the session to free up resources
	session.Close()

	// Reconnect to the keyspace
	cluster.Keyspace = keyspace
	session, err = cluster.CreateSession()
	if err != nil {
		log.Fatal("Unable to connect to keyspace:", err.Error())
	}
	defer session.Close()

}

func CreateSession() (*gocql.Session, error) {
	cassandraInstance := os.Getenv("CASSANDRA_INSTANCE")

	switch cassandraInstance {
	case "local":
		cluster := gocql.NewCluster(os.Getenv("CASSANDRA_HOST"))
		cassandraPort, err := strconv.Atoi(os.Getenv("CASSANDRA_PORT"))
		if err != nil {
			log.Fatal("Invalid Cassandra port:", err.Error())
		}
		cluster.Port = cassandraPort
		cluster.Keyspace = os.Getenv("CASSANDRA_URLLITE_KEYSPACE")
		cluster.Consistency = gocql.Quorum

		return cluster.CreateSession()
	default:
		cluster := gocql.NewCluster(os.Getenv("CASSANDRA_HOST"))
		cassandraPort, err := strconv.Atoi(os.Getenv("CASSANDRA_PORT"))
		if err != nil {
			log.Fatal("Invalid Cassandra port:", err.Error())
		}
		cluster.Port = cassandraPort
		cluster.Keyspace = os.Getenv("CASSANDRA_URLLITE_KEYSPACE")
		cluster.Consistency = gocql.Quorum

		return cluster.CreateSession()

	}

}
