// +build !jenkins

package brands

import (
	"encoding/json"
	"fmt"
	"github.com/Financial-Times/annotations-rw-neo4j/annotations"
	"github.com/Financial-Times/base-ft-rw-app-go/baseftrwapp"
	"github.com/Financial-Times/content-rw-neo4j/content"
	"github.com/Financial-Times/neo-utils-go/neoutils"
	"github.com/jmcvetta/neoism"
	"github.com/stretchr/testify/assert"
	"os"
	"sort"
	"testing"
	"reflect"
)

var defaultTypes = []string{"Thing", "Brand", "Concept", "Classification"}

const (
	validSkeletonBrandUuid = "92f4ec09-436d-4092-a88c-96f54e34007d"
	validSimpleBrandUuid   = "92f4ec09-436d-4092-a88c-96f54e34007c"
	validChildBrandUuid    = "a806e270-edbc-423f-b8db-d21ae90e06c8"
	specialCharBrandUuid   = "327af339-39d4-4c7b-8c06-9f80211ea93d"
	contentUuid            = "3fc9fe3e-af8c-4f7f-961a-e5065392bb31"
	parentBrandUuid = "92f4ec09-436d-4092-a88c-96f54e34007c"
)

var validSkeletonBrand = Brand{
	UUID:      validSkeletonBrandUuid,
	PrefLabel: "validSkeletonBrand",
	AlternativeIdentifiers: alternativeIdentifiers{
		TME:   []string{"111"},
		UUIDS: []string{"92f4ec09-436d-4092-a88c-96f54e34007d"},
	},
	Types: defaultTypes,
}

var updatedSkeletonBrand = Brand{
	UUID:      validSkeletonBrandUuid,
	PrefLabel: "validSkeletonBrand",
	AlternativeIdentifiers: alternativeIdentifiers{
		TME:   []string{"123"},
		UUIDS: []string{"92f4ec09-436d-4092-a88c-96f54e34007d"},
	},
	Types: defaultTypes,
}

var validSimpleBrand = Brand{
	UUID:           validSimpleBrandUuid,
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
	UUID:           validChildBrandUuid,
	ParentUUID:     parentBrandUuid,
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
	Aliases: []string{"SomeWonkyBrand", "AnotherAliasForABrand"},
}

var specialCharBrand = Brand{
	UUID:        specialCharBrandUuid,
	PrefLabel:   "specialCharBrand",
	Description: "This brand has a heart \u2665 and smiley \u263A",
	AlternativeIdentifiers: alternativeIdentifiers{
		TME:   []string{"1111"},
		UUIDS: []string{"327af339-39d4-4c7b-8c06-9f80211ea93d"},
	},
	Types: defaultTypes,
}

func TestCreateNotAllValuesPresent(t *testing.T) {
	assert := assert.New(t)
	db := getDatabaseConnectionAndCheckClean(t, assert)
	brandsDriver := getCypherDriver(db)

	defer cleanDB([]string{validSkeletonBrandUuid}, db, t, assert)

	assert.NoError(brandsDriver.Write(validSkeletonBrand, "TRANS"), "Failed to write brand")
	readBrandAndCompare(validSkeletonBrand, t, db)
}

func TestDeleteExistingBrandWithNoRelationshipsRemovesEverything(t *testing.T) {
	assert := assert.New(t)
	db := getDatabaseConnectionAndCheckClean(t, assert)
	brandsDriver := getCypherDriver(db)

	defer cleanDB([]string{validSimpleBrand.UUID}, db, t, assert)

	assert.NoError(brandsDriver.Write(validSimpleBrand, "TRANS"), "Failed to write brand")

	done, err := brandsDriver.Delete(validSimpleBrand.UUID, "TRANS")
	assert.NoError(err, "Failed to write brand")
	assert.True(done, "Delete failed to complete")

	brand, found, err := brandsDriver.Read(validSimpleBrand.UUID, "TRANS")
	assert.NoError(err, "Failed to read brand")
	assert.Equal(Brand{}, brand, "Found brand %s who should have been deleted", brand)
	assert.False(found, "Found a brand that should have been deleted")

	assert.False(doesThingExistAtAll(validSimpleBrand.UUID, db, t, assert), "Failed to delete brand")
}

func TestCreateAllValuesPresentAndParentNodeCreatedCorrectly(t *testing.T) {
	assert := assert.New(t)
	db := getDatabaseConnectionAndCheckClean(t, assert)
	brandsDriver := getCypherDriver(db)

	defer cleanDB([]string{parentBrandUuid, validChildBrand.UUID}, db, t, assert)

	err := brandsDriver.Write(validChildBrand, "TRANS")
	assert.NoError(err, "Failed to write brand")
	readBrandAndCompare(validChildBrand, t, db)

	assert.True(doesThingExistWithIdentifiers(parentBrandUuid, db, t, assert),
		"Unable to find a Thing with any Identifiers, uuid: %s", parentBrandUuid)

}

func TestCreateHandlesSpecialCharacters(t *testing.T) {
	assert := assert.New(t)
	db := getDatabaseConnectionAndCheckClean(t, assert)
	brandsDriver := getCypherDriver(db)

	defer cleanDB([]string{specialCharBrand.UUID}, db, t, assert)

	err := brandsDriver.Write(specialCharBrand, "TRANS")
	assert.NoError(err, "Failed to write brand")
	readBrandAndCompare(specialCharBrand, t, db)
}

func TestUpdateWillRemovePropertiesNoLongerPresent(t *testing.T) {
	assert := assert.New(t)
	db := getDatabaseConnectionAndCheckClean(t, assert)
	brandsDriver := getCypherDriver(db)

	defer cleanDB([]string{validSimpleBrand.UUID}, db, t, assert)

	myBrand := validSimpleBrand
	err := brandsDriver.Write(myBrand, "TRANS")
	readBrandAndCompare(myBrand, t, db)
	assert.NoError(err, "Failed to write brand")

	myBrand.Description = ""
	err = brandsDriver.Write(myBrand, "TRANS")
	assert.NoError(err, "Failed to write brand")
	readBrandAndCompare(myBrand, t, db)
}

func TestUpdateWillRemovePropertiesAndIdentifiersNoLongerPresent(t *testing.T) {
	assert := assert.New(t)
	db := getDatabaseConnectionAndCheckClean(t, assert)
	brandsDriver := getCypherDriver(db)

	defer cleanDB([]string{updatedSkeletonBrand.UUID, validSimpleBrand.UUID}, db, t, assert)

	assert.NoError(brandsDriver.Write(validSkeletonBrand, "TRANS"), "Failed to write brand")
	readBrandAndCompare(validSkeletonBrand, t, db)

	assert.NoError(brandsDriver.Write(updatedSkeletonBrand, "TRANS"), "Failed to write updated brand")
	readBrandAndCompare(updatedSkeletonBrand, t, db)
}

func TestCount(t *testing.T) {
	assert := assert.New(t)
	db := getDatabaseConnectionAndCheckClean(t, assert)
	brandsDriver := getCypherDriver(db)

	defer cleanDB([]string{updatedSkeletonBrand.UUID, validSimpleBrand.UUID}, db, t, assert)

	assert.NoError(brandsDriver.Write(validSkeletonBrand, "TRANS"), "Failed to write brand")

	nr, err := brandsDriver.Count()
	assert.Equal(1, nr, "Should be 1 subjects in DB - count differs")
	assert.NoError(err, "An unexpected error occurred during count")

	assert.NoError(brandsDriver.Write(validSimpleBrand, "TRANS"), "Failed to write brand")

	nr, err = brandsDriver.Count()
	assert.Equal(2, nr, "Should be 2 subjects in DB - count differs")
	assert.NoError(err, "An unexpected error occurred during count")
}

func TestConnectivityCheck(t *testing.T) {
	assert := assert.New(t)
	db := getDatabaseConnectionAndCheckClean(t, assert)
	brandsDriver := getCypherDriver(db)

	err := brandsDriver.Check()
	assert.NoError(err, "Check connectivity failed")
}

func TestDeleteWithRelationshipsMaintainsRelationships(t *testing.T) {
	assert := assert.New(t)
	db := getDatabaseConnectionAndCheckClean(t, assert)
	brandsDriver := getCypherDriver(db)

	defer cleanDB([]string{validSimpleBrandUuid, contentUuid}, db, t, assert)

	assert.NoError(brandsDriver.Write(validSimpleBrand, "TRANS"), "Failed to write brand")
	writeContent(assert, db)
	writeAnnotation(assert, db)

	found, err := brandsDriver.Delete(validSimpleBrandUuid, "TRANS")

	assert.True(found, "Didn't manage to delete brand for uuid %", validSimpleBrandUuid)
	assert.NoError(err, "Error deleting brand for uuid %s", validSimpleBrandUuid)

	brand, found, err := brandsDriver.Read(validSimpleBrandUuid, "TRANS")

	assert.Equal(Brand{}, brand, "Found brand %s who should have been deleted", brand)
	assert.False(found, "Found brand for uuid %s who should have been deleted", validSimpleBrandUuid)
	assert.NoError(err, "Error trying to find brand for uuid %s", validSimpleBrandUuid)
	assert.True(doesThingExistWithIdentifiers(validSimpleBrandUuid, db, t, assert),
		"Unable to find a Thing with any Identifiers, uuid: %s", validSimpleBrandUuid)
}

func writeAnnotation(assert *assert.Assertions, db neoutils.NeoConnection) annotations.Service {
	annotationsRW := annotations.NewCypherAnnotationsService(db, "v2", "TRANS")
	assert.NoError(annotationsRW.Initialise())
	writeJSONToAnnotationsService(annotationsRW, contentUuid, "./fixtures/Annotations-3fc9fe3e-af8c-4f7f-961a-e5065392bb31-v2.json", assert)
	return annotationsRW
}

func writeContent(assert *assert.Assertions, db neoutils.NeoConnection) baseftrwapp.Service {
	contentRW := content.NewCypherContentService(db)
	assert.NoError(contentRW.Initialise())
	writeJSONToService(contentRW, "./fixtures/Content-3fc9fe3e-af8c-4f7f-961a-e5065392bb31.json", assert)
	return contentRW
}

func writeJSONToAnnotationsService(service annotations.Service, contentUUID string, pathToJSONFile string, assert *assert.Assertions) {
	f, err := os.Open(pathToJSONFile)
	assert.NoError(err)
	dec := json.NewDecoder(f)
	inst, errr := service.DecodeJSON(dec)
	assert.NoError(errr, "Error parsing file %s", pathToJSONFile)
	errrr := service.Write(contentUUID, inst)
	assert.NoError(errrr)
}

func writeJSONToService(service baseftrwapp.Service, pathToJSONFile string, assert *assert.Assertions) {
	f, err := os.Open(pathToJSONFile)
	assert.NoError(err)
	dec := json.NewDecoder(f)
	inst, _, errr := service.DecodeJSON(dec)
	assert.NoError(errr)
	errrr := service.Write(inst, "TRANS")
	assert.NoError(errrr)
}

func readBrandAndCompare(expected Brand, t *testing.T, db neoutils.NeoConnection) {
	sort.Strings(expected.Types)
	sort.Strings(expected.Aliases)

	actual, found, err := getCypherDriver(db).Read(expected.UUID, "TRANS")
	assert.NoError(t, err, "Failed to read brand")
	assert.True(t, found, "Failed to find brand")

	actualBrand := actual.(Brand)
	sort.Strings(actualBrand.Types)
	sort.Strings(actualBrand.Aliases)
	assert.True(t, reflect.DeepEqual(expected, actual), "Actual brand and exepected brand do not match")
}

func getCypherDriver(db neoutils.NeoConnection) service {
	cr := NewCypherBrandsService(db)
	cr.Initialise()
	return cr
}

func getDatabaseConnectionAndCheckClean(t *testing.T, assert *assert.Assertions) neoutils.NeoConnection {
	db := getDatabaseConnection(assert)
	checkDbClean([]string{
		validSkeletonBrandUuid,
		validSimpleBrandUuid,
		validChildBrandUuid,
		specialCharBrandUuid,
		contentUuid,
		parentBrandUuid}, db, t)
	return db
}

func getDatabaseConnection(assert *assert.Assertions) neoutils.NeoConnection {
	url := os.Getenv("NEO4J_TEST_URL")
	if url == "" {
		url = "http://localhost:7474/db/data"
	}

	conf := neoutils.DefaultConnectionConfig()
	conf.Transactional = false
	db, err := neoutils.Connect(url, conf)
	assert.NoError(err, "Failed to connect to Neo4j")
	return db
}

func cleanDB(uuidsToClean []string, db neoutils.NeoConnection, t *testing.T, assert *assert.Assertions) {
	qs := make([]*neoism.CypherQuery, len(uuidsToClean))
	for i, uuid := range uuidsToClean {
		qs[i] = &neoism.CypherQuery{
			Statement: fmt.Sprintf(`
			MATCH (a:Thing {uuid: "%s"})
			OPTIONAL MATCH (a)-[rel]-(s)
			OPTIONAL MATCH (p)-[rel2]-(id)
			DELETE rel2, id, rel, s
			DETACH DELETE a`, uuid)}
	}

	err := db.CypherBatch(qs)
	assert.NoError(err)
}

func checkDbClean(uuidsCleaned []string, db neoutils.NeoConnection, t *testing.T) {
	assert := assert.New(t)

	result := []struct {
		Uuid string `json:"thing.uuid"`
	}{}

	checkGraph := neoism.CypherQuery{
		Statement: `
			MATCH (thing) WHERE thing.uuid in {uuids} RETURN thing.uuid
		`,
		Parameters: neoism.Props{
			"uuids": uuidsCleaned,
		},
		Result: &result,
	}
	err := db.CypherBatch([]*neoism.CypherQuery{&checkGraph})
	assert.NoError(err)
	assert.Empty(result)
}

func doesThingExistAtAll(uuid string, db neoutils.NeoConnection, t *testing.T, assert *assert.Assertions) bool {
	result := []struct {
		Uuid string `json:"thing.uuid"`
	}{}

	checkGraph := neoism.CypherQuery{
		Statement: `
			MATCH (a:Thing {uuid: "%s"}) return a.uuid
		`,
		Parameters: neoism.Props{
			"uuid": uuid,
		},
		Result: &result,
	}

	err := db.CypherBatch([]*neoism.CypherQuery{&checkGraph})
	assert.NoError(err)

	if len(result) == 0 {
		return false
	}

	return true
}

func doesThingExistWithIdentifiers(uuid string, db neoutils.NeoConnection, t *testing.T, assert *assert.Assertions) bool {

	result := []struct {
		uuid string `json:"thing.uuid"`
	}{}

	checkGraph := neoism.CypherQuery{
		Statement: `
			MATCH (a:Thing {uuid: "%s"})-[:IDENTIFIES]-(:Identifier)
			WITH collect(distinct a.uuid) as uuid
			RETURN uuid
		`,
		Parameters: neoism.Props{
			"uuid": uuid,
		},
		Result: &result,
	}

	err := db.CypherBatch([]*neoism.CypherQuery{&checkGraph})
	assert.NoError(err)

	if len(result) == 0 {
		return false
	}

	fmt.Printf("Result3: %v", len(result))
	return true
}
