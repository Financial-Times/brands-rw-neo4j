// +build !jenkins

package brands

import (
	"os"
	"testing"

	"github.com/Financial-Times/base-ft-rw-app-go"
	"github.com/Financial-Times/neo-utils-go"
	"github.com/jmcvetta/neoism"
	"github.com/stretchr/testify/assert"
)

var peopleDriver baseftrwapp.Service

func TestDelete(t *testing.T) {
	assert := assert.New(t)
	uuid := "12345"

	peopleDriver = getPeopleCypherDriver(t)

	personToDelete := person{UUID: uuid, Name: "Test", BirthYear: 1974, Salutation: "Mr",
		Identifiers: []identifier{identifier{fsAuthority, "FACTSET_ID"}}}

	assert.NoError(peopleDriver.Write(personToDelete), "Failed to write person")

	found, err := peopleDriver.Delete(uuid)
	assert.True(found, "Didn't manage to delete person for uuid %", uuid)
	assert.NoError(err, "Error deleting person for uuid %s", uuid)

	p, found, err := peopleDriver.Read(uuid)

	assert.Equal(person{}, p, "Found person %s who should have been deleted", p)
	assert.False(found, "Found person for uuid %s who should have been deleted", uuid)
	assert.NoError(err, "Error trying to find person for uuid %s", uuid)
}

func TestCreateAllValuesPresent(t *testing.T) {
	assert := assert.New(t)
	uuid := "12345"
	peopleDriver = getPeopleCypherDriver(t)

	personToWrite := person{UUID: uuid, Name: "Test", BirthYear: 1974, Salutation: "Mr",
		Identifiers: []identifier{identifier{fsAuthority, "FACTSET_ID"}}}

	assert.NoError(peopleDriver.Write(personToWrite), "Failed to write person")

	readPersonForUUIDAndCheckFieldsMatch(t, uuid, personToWrite)

	cleanUp(t, uuid)
}

func TestCreateHandlesSpecialCharacters(t *testing.T) {
	assert := assert.New(t)
	uuid := "12345"
	peopleDriver = getPeopleCypherDriver(t)

	personToWrite := person{UUID: uuid, Name: "Thomas M. O'Gara", BirthYear: 1974, Salutation: "Mr",
		Identifiers: []identifier{identifier{fsAuthority, "FACTSET_ID"}}}

	assert.NoError(peopleDriver.Write(personToWrite), "Failed to write person")

	readPersonForUUIDAndCheckFieldsMatch(t, uuid, personToWrite)

	cleanUp(t, uuid)
}

func TestCreateNotAllValuesPresent(t *testing.T) {
	assert := assert.New(t)
	uuid := "12345"
	peopleDriver = getPeopleCypherDriver(t)

	personToWrite := person{UUID: uuid, Name: "Test",
		Identifiers: []identifier{identifier{fsAuthority, "FACTSET_ID"}}}

	assert.NoError(peopleDriver.Write(personToWrite), "Failed to write person")

	readPersonForUUIDAndCheckFieldsMatch(t, uuid, personToWrite)

	cleanUp(t, uuid)
}

func TestUpdateWillRemovePropertiesNoLongerPresent(t *testing.T) {
	assert := assert.New(t)
	uuid := "12345"
	peopleDriver = getPeopleCypherDriver(t)

	personToWrite := person{UUID: uuid, Name: "Test", BirthYear: 1974, Salutation: "Mr",
		Identifiers: []identifier{identifier{fsAuthority, "FACTSET_ID"}}}

	assert.NoError(peopleDriver.Write(personToWrite), "Failed to write person")
	readPersonForUUIDAndCheckFieldsMatch(t, uuid, personToWrite)

	updatedPerson := person{UUID: uuid, Name: "Test",
		Identifiers: []identifier{identifier{fsAuthority, "FACTSET_ID"}}}

	assert.NoError(peopleDriver.Write(updatedPerson), "Failed to write updated person")
	readPersonForUUIDAndCheckFieldsMatch(t, uuid, updatedPerson)

	cleanUp(t, uuid)
}

func TestConnectivityCheck(t *testing.T) {
	assert := assert.New(t)
	peopleDriver = getPeopleCypherDriver(t)
	err := peopleDriver.Check()
	assert.NoError(err, "Unexpected error on connectivity check")
}

func getPeopleCypherDriver(t *testing.T) service {
	assert := assert.New(t)
	url := os.Getenv("NEO4J_TEST_URL")
	if url == "" {
		url = "http://localhost:7474/db/data"
	}

	db, err := neoism.Connect(url)
	assert.NoError(err, "Failed to connect to Neo4j")
	return NewCypherPeopleService(neoutils.StringerDb{db}, db)
}

func readPersonForUUIDAndCheckFieldsMatch(t *testing.T, uuid string, expectedPerson person) {
	assert := assert.New(t)
	storedPerson, found, err := peopleDriver.Read(uuid)

	assert.NoError(err, "Error finding person for uuid %s", uuid)
	assert.True(found, "Didn't find person for uuid %s", uuid)
	assert.Equal(expectedPerson, storedPerson, "people should be the same")
}

func cleanUp(t *testing.T, uuid string) {
	assert := assert.New(t)
	found, err := peopleDriver.Delete(uuid)
	assert.True(found, "Didn't manage to delete person for uuid %", uuid)
	assert.NoError(err, "Error deleting person for uuid %s", uuid)
}
