// +build !jenkins

package brands

import (
	"github.com/Financial-Times/base-ft-rw-app-go/baseftrwapp"
	"github.com/Financial-Times/neo-utils-go/neoutils"
	"github.com/jmcvetta/neoism"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

var defaultTypes = []string{"Thing", "Brand", "Concept", "Classification"}

var validSkeletonBrand = Brand{
	UUID:      "92f4ec09-436d-4092-a88c-96f54e34007d",
	PrefLabel: "validSkeletonBrand",
	AlternativeIdentifiers: alternativeIdentifiers{
			TME: []string{"111"},
			UUIDS: []string{"92f4ec09-436d-4092-a88c-96f54e34007d"},
			FactsetIdentifier: "fsetId",
	},
	Types:defaultTypes,
}

var updatedSkeletonBrand = Brand{
	UUID:      "92f4ec09-436d-4092-a88c-96f54e34007d",
	PrefLabel: "validSkeletonBrand",
	AlternativeIdentifiers: alternativeIdentifiers{
		TME: []string{"123"},
		UUIDS: []string{"92f4ec09-436d-4092-a88c-96f54e34007d"},
	},
	Types:defaultTypes,
}

var validSimpleBrand = Brand{
	UUID:           "92f4ec09-436d-4092-a88c-96f54e34007c",
	PrefLabel:      "validSimpleBrand",
	Strapline:      "Keeping it simple",
	Description:    "This brand has no parent but otherwise has valid values for all fields",
	DescriptionXML: "<body>This <i>brand</i> has no parent but otherwise has valid values for all fields</body>",
	ImageURL:       "http://media.ft.com/validSimpleBrand.png",
	AlternativeIdentifiers: alternativeIdentifiers{
			TME: []string{"123"},
			UUIDS: []string{"92f4ec09-436d-4092-a88c-96f54e34007c"},
	},
	Types:defaultTypes,
}

var validChildBrand = Brand{
	UUID:           "a806e270-edbc-423f-b8db-d21ae90e06c8",
	ParentUUID:     "92f4ec09-436d-4092-a88c-96f54e34007c",
	PrefLabel:      "validChildBrand",
	Strapline:      "My parent is simple",
	Description:    "This brand has a parent and valid values for all fields",
	DescriptionXML: "<body>This <i>brand</i> has a parent and valid values for all fields</body>",
	ImageURL:       "http://media.ft.com/validChildBrand.png",
	AlternativeIdentifiers: alternativeIdentifiers{
			TME: []string{"123123"},
			UUIDS: []string{"a806e270-edbc-423f-b8db-d21ae90e06c8"},
	},
	Types:defaultTypes,
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
		AlternativeIdentifiers: alternativeIdentifiers{
			TME: []string{"1111"},
			UUIDS: []string{"327af339-39d4-4c7b-8c06-9f80211ea93d"},
		},
		Types:defaultTypes,
	}
	err := getCypherDriver(t).Write(specialCharBrand)
	assert.NoError(t, err)
	readBrandAndCompare(specialCharBrand, t)
	cleanUp(specialCharBrand.UUID, t)
}

func TestUpdateWillRemovePropertiesNoLongerPresent(t *testing.T) {
	myBrand := validSimpleBrand
	err := getCypherDriver(t).Write(myBrand)
	readBrandAndCompare(myBrand, t)
	assert.NoError(t, err)
	myBrand.Description = ""
	err = getCypherDriver(t).Write(myBrand)
	assert.NoError(t, err)
	readBrandAndCompare(myBrand, t)
	cleanUp(myBrand.UUID, t)
}

func TestUpdateWillRemovePropertiesAndIdentifiersNoLongerPresent(t *testing.T) {
	assert := assert.New(t)
	brandsDriver := getCypherDriver(t)

	assert.NoError(brandsDriver.Write(validSkeletonBrand), "Failed to write brand")
	readBrandAndCompare(validSkeletonBrand, t)

	assert.NoError(brandsDriver.Write(updatedSkeletonBrand), "Failed to write updated brand")
	readBrandAndCompare(updatedSkeletonBrand, t)

	cleanUp(updatedSkeletonBrand.UUID, t)
}

func TestCount(t *testing.T) {
	assert := assert.New(t)
	brandsDriver := getCypherDriver(t)

	assert.NoError(brandsDriver.Write(validSkeletonBrand), "Failed to write brand")

	nr, err := brandsDriver.Count()
	assert.Equal(1, nr, "Should be 1 subjects in DB - count differs")
	assert.NoError(err, "An unexpected error occurred during count")

	assert.NoError(brandsDriver.Write(validSimpleBrand), "Failed to write brand")

	nr, err = brandsDriver.Count()
	assert.Equal(2, nr, "Should be 2 subjects in DB - count differs")
	assert.NoError(err, "An unexpected error occurred during count")

	cleanUp(validSkeletonBrand.UUID, t)
	cleanUp(validSimpleBrand.UUID, t)
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
