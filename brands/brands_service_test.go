// +build !jenkins

package brands

import (
	"github.com/Financial-Times/neo-utils-go/neoutils"
	"github.com/stretchr/testify/assert"
	"os"
	"sort"
	"testing"
)

var defaultTypes = []string{"Thing", "Brand", "Concept", "Classification"}

var validSkeletonBrand = Brand{
	UUID:      "92f4ec09-436d-4092-a88c-96f54e34007d",
	PrefLabel: "validSkeletonBrand",
	AlternativeIdentifiers: alternativeIdentifiers{
		TME:   []string{"111"},
		UUIDS: []string{"92f4ec09-436d-4092-a88c-96f54e34007d"},
	},
	Types: defaultTypes,
}

var updatedSkeletonBrand = Brand{
	UUID:      "92f4ec09-436d-4092-a88c-96f54e34007d",
	PrefLabel: "validSkeletonBrand",
	AlternativeIdentifiers: alternativeIdentifiers{
		TME:   []string{"123"},
		UUIDS: []string{"92f4ec09-436d-4092-a88c-96f54e34007d"},
	},
	Types: defaultTypes,
}

var validSimpleBrand = Brand{
	UUID:           "92f4ec09-436d-4092-a88c-96f54e34007c",
	PrefLabel:      "validSimpleBrand",
	Strapline:      "Keeping it simple",
	Description:    "This brand has no parent but otherwise has valid values for all fields",
	DescriptionXML: "<body>This <i>brand</i> has no parent but otherwise has valid values for all fields</body>",
	ImageURL:       "http://media.ft.com/validSimpleBrand.png",
	AlternativeIdentifiers: alternativeIdentifiers{
		TME:   []string{"123"},
		UUIDS: []string{"92f4ec09-436d-4092-a88c-96f54e34007c"},
	},
	Types: defaultTypes,
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
		TME:   []string{"123123"},
		UUIDS: []string{"a806e270-edbc-423f-b8db-d21ae90e06c8"},
	},
	Types: defaultTypes,
}

func TestCreateNotAllValuesPresent(t *testing.T) {
	err := getCypherDriver(getDatabaseConnection(t)).Write(validSkeletonBrand)
	assert.NoError(t, err)
	readBrandAndCompare(validSkeletonBrand, t)
	cleanUp(validSkeletonBrand.UUID, t)
}

func TestDeleteExistingBrand(t *testing.T) {
	driver := getCypherDriver(getDatabaseConnection(t))
	err := driver.Write(validSimpleBrand)
	assert.NoError(t, err)

	done, err := getCypherDriver(getDatabaseConnection(t)).Delete(validSimpleBrand.UUID)
	assert.NoError(t, err)
	assert.True(t, done)

	person, found, err := getCypherDriver(getDatabaseConnection(t)).Read(validSimpleBrand.UUID)
	assert.NoError(t, err)
	assert.EqualValues(t, Brand{}, person)
	assert.False(t, found)
}

func TestCreateAllValuesPresent(t *testing.T) {
	err := getCypherDriver(getDatabaseConnection(t)).Write(validChildBrand)
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
			TME:   []string{"1111"},
			UUIDS: []string{"327af339-39d4-4c7b-8c06-9f80211ea93d"},
		},
		Types: defaultTypes,
	}
	err := getCypherDriver(getDatabaseConnection(t)).Write(specialCharBrand)
	assert.NoError(t, err)
	readBrandAndCompare(specialCharBrand, t)
	cleanUp(specialCharBrand.UUID, t)
}

func TestUpdateWillRemovePropertiesNoLongerPresent(t *testing.T) {
	myBrand := validSimpleBrand
	err := getCypherDriver(getDatabaseConnection(t)).Write(myBrand)
	readBrandAndCompare(myBrand, t)
	assert.NoError(t, err)
	myBrand.Description = ""
	err = getCypherDriver(getDatabaseConnection(t)).Write(myBrand)
	assert.NoError(t, err)
	readBrandAndCompare(myBrand, t)
	cleanUp(myBrand.UUID, t)
}

func TestUpdateWillRemovePropertiesAndIdentifiersNoLongerPresent(t *testing.T) {
	brandsDriver := getCypherDriver(getDatabaseConnection(t))

	assert.NoError(t, brandsDriver.Write(validSkeletonBrand), "Failed to write brand")
	readBrandAndCompare(validSkeletonBrand, t)

	assert.NoError(t, brandsDriver.Write(updatedSkeletonBrand), "Failed to write updated brand")
	readBrandAndCompare(updatedSkeletonBrand, t)

	cleanUp(updatedSkeletonBrand.UUID, t)
}

func TestCount(t *testing.T) {
	brandsDriver := getCypherDriver(getDatabaseConnection(t))

	assert.NoError(t, brandsDriver.Write(validSkeletonBrand), "Failed to write brand")

	nr, err := brandsDriver.Count()
	assert.Equal(t, 1, nr, "Should be 1 subjects in DB - count differs")
	assert.NoError(t, err, "An unexpected error occurred during count")

	assert.NoError(t, brandsDriver.Write(validSimpleBrand), "Failed to write brand")

	nr, err = brandsDriver.Count()
	assert.Equal(t, 2, nr, "Should be 2 subjects in DB - count differs")
	assert.NoError(t, err, "An unexpected error occurred during count")

	cleanUp(validSkeletonBrand.UUID, t)
	cleanUp(validSimpleBrand.UUID, t)
}

func TestConnectivityCheck(t *testing.T) {
	driver := getCypherDriver(getDatabaseConnection(t))
	err := driver.Check()
	assert.NoError(t, err)
}

func getDatabaseConnection(t *testing.T) neoutils.NeoConnection {
	url := os.Getenv("NEO4J_TEST_URL")
	if url == "" {
		url = "http://localhost:7474/db/data"
	}

	conf := neoutils.DefaultConnectionConfig()
	conf.Transactional = false
	db, err := neoutils.Connect(url, conf)
	assert.NoError(t, err, "Failed to connect to Neo4j")
	return db
}

func getCypherDriver(db neoutils.NeoConnection) service {
	cr := NewCypherBrandsService(db)
	cr.Initialise()
	return cr
}

func readBrandAndCompare(expected Brand, t *testing.T) {
	sort.Strings(expected.Types)

	actual, found, err := getCypherDriver(getDatabaseConnection(t)).Read(expected.UUID)
	assert.NoError(t, err)
	assert.True(t, found)

	actualBrand := actual.(Brand)
	sort.Strings(actualBrand.Types)
	assert.EqualValues(t, expected, actualBrand)
}

func cleanUp(uuid string, t *testing.T) {
	found, err := getCypherDriver(getDatabaseConnection(t)).Delete(uuid)
	assert.True(t, found, "Didn't manage to delete brand for uuid %s", uuid)
	assert.NoError(t, err, "Error deleting brand for uuid %s", uuid)
}
