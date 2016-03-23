// +build !jenkins

package brands

import (
	//"fmt"
	"github.com/Financial-Times/base-ft-rw-app-go/baseftrwapp"
	"github.com/Financial-Times/neo-utils-go/neoutils"
	"github.com/jmcvetta/neoism"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

var validSkeletonBrand = Brand{
	UUID:      "92f4ec09-436d-4092-a88c-96f54e34007d",
	PrefLabel: "validSkeletonBrand",
	Identifiers: []identifier{
		identifier{
			Authority:       tmeAuthority,
			IdentifierValue: "111",
		},
	},
}

var validSimpleBrand = Brand{
	UUID:           "92f4ec09-436d-4092-a88c-96f54e34007c",
	PrefLabel:      "validSimpleBrand",
	Strapline:      "Keeping it simple",
	Description:    "This brand has no parent but otherwise has valid values for all fields",
	DescriptionXML: "<body>This <i>brand</i> has no parent but otherwise has valid values for all fields</body>",
	ImageURL:       "http://media.ft.com/validSimpleBrand.png",
	Identifiers: []identifier{
		identifier{
			Authority:       tmeAuthority,
			IdentifierValue: "123",
		},
	},
}

var validChildBrand = Brand{
	UUID:           "a806e270-edbc-423f-b8db-d21ae90e06c8",
	ParentUUID:     "92f4ec09-436d-4092-a88c-96f54e34007c",
	PrefLabel:      "validChildBrand",
	Strapline:      "My parent is simple",
	Description:    "This brand has a parent and valid values for all fields",
	DescriptionXML: "<body>This <i>brand</i> has a parent and valid values for all fields</body>",
	ImageURL:       "http://media.ft.com/validChildBrand.png",
	Identifiers: []identifier{
		identifier{
			Authority:       tmeAuthority,
			IdentifierValue: "123123",
		},
	},
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
	assert.NoError(t, err)
	assert.True(t, done)

	person, found, err := getCypherDriver(t).Read(validSimpleBrand.UUID)
	assert.NoError(t, err)
	assert.EqualValues(t, Brand{}, person)
	assert.False(t, found)
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
		Identifiers: []identifier{
			identifier{
				Authority:       tmeAuthority,
				IdentifierValue: "1111",
			},
		},
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
	//fmt.Printf("Looking for %+v\n", expected.UUID)
	actual, found, err := getCypherDriver(t).Read(expected.UUID)
	assert.NoError(t, err)
	assert.True(t, found)
	//fmt.Printf("Found %+v\n", actual)
	assert.EqualValues(t, expected, actual)
}

func cleanUp(uuid string, t *testing.T) {
	found, err := getCypherDriver(t).Delete(uuid)
	assert.True(t, found, "Didn't manage to delete brand for uuid %s", uuid)
	assert.NoError(t, err, "Error deleting brand for uuid %s", uuid)
}
