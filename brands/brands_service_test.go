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

var validOrphanBrand = Brand{
	UUID:           "92f4ec09-436d-4092-a88c-96f54e34007c",
	PrefLabel:      "validOrphanBrand",
	Description:    "This brand has no parent but otherwise has valid values for all fields",
	DescriptionXML: "<body>This <i>brand</i> has no parent but otherwise has valid values for all fields</body>",
	ImageURL:       "http://media.ft.com/validFatOrphanBrand.png",
}

func TestDeleteExitingBrand(t *testing.T) {
	driver := getCypherDriver(t)
	err := driver.Write(validOrphanBrand)
	assert.NoError(t, err)
	done, err := getCypherDriver(t).Delete(validOrphanBrand.UUID)
	assert.True(t, done)
}

func TestCreateAllValuesPresent(t *testing.T) {
	err := getCypherDriver(t).Write(validOrphanBrand)
	assert.NoError(t, err)
	readBrandAndCompare(validOrphanBrand, t)
	cleanUp(validOrphanBrand.UUID, t)
}

func readBrandAndCompare(expected Brand, t *testing.T) {
	actual, found, err := getCypherDriver(t).Read(expected.UUID)
	assert.NoError(t, err)
	assert.True(t, found)
	assert.EqualValues(t, expected, actual)
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
	err := driver.Check()
	assert.NoError(t, err)
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

func cleanUp(uuid string, t *testing.T) {
	found, err := getCypherDriver(t).Delete(uuid)
	assert.True(t, found, "Didn't manage to delete brand for uuid %", uuid)
	assert.NoError(t, err, "Error deleting brand for uuid %s", uuid)
}
