// +build !jenkins

package brands

import (
	"github.com/Financial-Times/base-ft-rw-app-go"
	"github.com/Financial-Times/neo-utils-go"
	"github.com/jmcvetta/neoism"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestDelete(t *testing.T) {
	assert := assert.New(t)
	uuid := "123"

	driver := getCypherDriver(t)

	instData := Brand{UUID: uuid, PrefLabel: "Test"}
	assert.NoError(driver.Write(instData), "Error creating test instance data with uuid %s", uuid)

	found, err := driver.Delete(uuid)
	assert.NoError(err, "Error deleting test instance data for uuid %s", uuid)
	assert.True(found, "Unable to delete test instance data with uuid %s (%+v)", uuid, instData)

	zombieInst, found, err := driver.Read(uuid)

	assert.Equal(Brand{}, zombieInst,
		"Test instance data with uuid %s should have been deleted (found %+v)", uuid, zombieInst)
	assert.False(found, "Found instance with uuid %s, should have been deleted", uuid)
	assert.NoError(err, "Error looking instance with uuid %s", uuid)
}

func TestCreateAllValuesPresent(t *testing.T) {
	// cleanUp(t, uuid)
}

func TestCreateHandlesSpecialCharacters(t *testing.T) {
}

func TestCreateNotAllValuesPresent(t *testing.T) {
	// cleanUp(t, uuid)
}

func TestUpdateWillRemovePropertiesNoLongerPresent(t *testing.T) {
	// readPersonForUUIDAndCheckFieldsMatch(t, uuid, personToWrite)
	//
	// readPersonForUUIDAndCheckFieldsMatch(t, uuid, updatedPerson)
	//
	// cleanUp(t, uuid)
}

func TestConnectivityCheck(t *testing.T) {
	driver := getCypherDriver(t)
	assert := assert.New(t)
	err := driver.Check()
	assert.NoError(err, "Unexpected error on connectivity check")
}

func getCypherDriver(t *testing.T) (service baseftrwapp.Service) {
	url := os.Getenv("NEO4J_TEST_URL")
	if url == "" {
		url = "http://localhost:7474/db/data"
	}
	db, err := neoism.Connect(url)
	assert.New(t).NoError(err, "Error setting up connection to %s", url)
	return NewCypherBrandsService(neoutils.StringerDb{db}, db)
}

func readInstanceByUUIDAndCheckFieldsMatch(driver *baseftrwapp.Service, t *testing.T, uuid string, expectedData interface{}) {
	assert := assert.New(t)
	storedData, found, err := driver.Read(uuid)
	assert.NoError(err, "Error finding instance data for uuid %s", uuid)
	assert.True(found, "Didn't find instance data for uuid %s", uuid)
	assert.EqualValues(expectedData, storedData, "instance data should be the same")
}

func cleanUp(t *testing.T, uuid string) {
	assert := assert.New(t)
	found, err := driver.Delete(uuid)
	assert.True(found, "Didn't manage to delete brand for uuid %", uuid)
	assert.NoError(err, "Error deleting brand for uuid %s", uuid)
}
