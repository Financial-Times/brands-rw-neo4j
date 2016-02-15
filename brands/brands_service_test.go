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

var validSkeletonBrand = Brand{
	UUID:      "92f4ec09-436d-4092-a88c-96f54e34007c",
	PrefLabel: "validSkeletonBrand",
}

var validSimpleBrand = Brand{
	UUID:           "92f4ec09-436d-4092-a88c-96f54e34007c",
	PrefLabel:      "validSimpleBrand",
	Strapline:      "Keeping it simple",
	Description:    "This brand has no parent but otherwise has valid values for all fields",
	DescriptionXML: "<body>This <i>brand</i> has no parent but otherwise has valid values for all fields</body>",
	ImageURL:       "http://media.ft.com/validSimpleBrand.png",
}

var validChildBrand = Brand{
	UUID:           "a806e270-edbc-423f-b8db-d21ae90e06c8",
	ParentUUID:     "92f4ec09-436d-4092-a88c-96f54e34007c",
	PrefLabel:      "validChildBrand",
	Strapline:      "My parent is simple",
	Description:    "This brand has a parent and valid values for all fields",
	DescriptionXML: "<body>This <i>brand</i> has a parent and valid values for all fields</body>",
	ImageURL:       "http://media.ft.com/validChildBrand.png",
}

func TestCreateNotAllValuesPresent(t *testing.T) {
	err := getCypherDriver(t).Write(validSkeletonBrand)
	assert.NoError(t, err)
	readBrandAndCompare(validSkeletonBrand, t)
	cleanUp(validSkeletonBrand.UUID, t)
}

func TestDeleteExistingBrand(t *testing.T) {
	driver := getCypherDriver(t)
	err := driver.Write(validSimpleBrand)
	assert.NoError(t, err)
	done, err := getCypherDriver(t).Delete(validSimpleBrand.UUID)
	assert.True(t, done)
}

func TestCreateAllValuesPresent(t *testing.T) {
	err := getCypherDriver(t).Write(validChildBrand)
	assert.NoError(t, err)
	readBrandAndCompare(validChildBrand, t)
	cleanUp(validChildBrand.UUID, t)
}

func TestCreateHandlesSpecialCharacters(t *testing.T) {
	specialCharBrand := Brand{
		UUID:        "327af339-39d4-4c7b-8c06-9f80211ea93d",
		PrefLabel:   "specialCharBrand",
		Description: "This brand has a heart \u2665 and smiley \u263A",
	}
	err := getCypherDriver(t).Write(specialCharBrand)
	assert.NoError(t, err)
	readBrandAndCompare(specialCharBrand, t)
	cleanUp(specialCharBrand.UUID, t)
}

func TestUpdateWillRemovePropertiesNoLongerPresent(t *testing.T) {
	err := getCypherDriver(t).Write(validSimpleBrand)
	assert.NoError(t, err)
	err = getCypherDriver(t).Write(validSkeletonBrand)
	assert.NoError(t, err)
	readBrandAndCompare(validSkeletonBrand, t)
	cleanUp(validSkeletonBrand.UUID, t)
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
	assert.NoError(t, err, "Error setting up connection to %s", url)
	return NewCypherBrandsService(neoutils.StringerDb{db}, db)
}

func readBrandAndCompare(expected Brand, t *testing.T) {
	actual, found, err := getCypherDriver(t).Read(expected.UUID)
	assert.NoError(t, err)
	assert.True(t, found)
	assert.EqualValues(t, expected, actual)
}

func cleanUp(uuid string, t *testing.T) {
	found, err := getCypherDriver(t).Delete(uuid)
	assert.True(t, found, "Didn't manage to delete brand for uuid %s", uuid)
	assert.NoError(t, err, "Error deleting brand for uuid %s", uuid)
}
