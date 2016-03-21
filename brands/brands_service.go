package brands

import (
	"encoding/json"
	"fmt"
	"github.com/Financial-Times/neo-utils-go/neoutils"
	log "github.com/Sirupsen/logrus"
	"github.com/jmcvetta/neoism"
)

//service maintains info about runners and index managers
type service struct {
	cypherRunner neoutils.CypherRunner
	indexManager neoutils.IndexManager
}

// NewCypherBrandsService provides functions for create, update, delete operations on brands in Neo4j,
// plus other utility functions needed for a service
func NewCypherBrandsService(cypherRunner neoutils.CypherRunner, indexManager neoutils.IndexManager) service {
	return service{cypherRunner, indexManager}
}

//Initialise the driver
func (s service) Initialise() error {
	entities := map[string]string{
		"Thing":      "uuid",
		"Concept":    "uuid",
		"Brand":      "uuid",
		"Identifier": "value",
	}
	if err := neoutils.EnsureConstraints(s.indexManager, entities); err != nil {
		return err
	}
	if err := neoutils.EnsureIndexes(s.indexManager, entities); err != nil {
		return err
	}
	return nil
}

func (s service) Read(uuid string) (interface{}, bool, error) {
	results := []struct {
		Brand
	}{}
	query := &neoism.CypherQuery{
		Statement: `
                        MATCH (n:Brand {uuid:{uuid}})
                        OPTIONAL MATCH (n:Brand {uuid:{uuid}})-[:HAS_PARENT]->(p:Thing)
                        OPTIONAL MATCH (n)<-[:IDENTIFIES]-(i:Identifier)
                        RETURN n.uuid AS uuid, n.prefLabel AS prefLabel,
                                n.strapline AS strapline, p.uuid as parentUUID,
                                n.descriptionXML AS descriptionXML,
                                n.description AS description, n.imageUrl AS _imageUrl,
                                collect({authority:i.authority, identifierValue:i.value}) as identifiers
                                `,
		Parameters: map[string]interface{}{
			"uuid": uuid,
		},
		Result: &results,
	}
	err := s.cypherRunner.CypherBatch([]*neoism.CypherQuery{query})
	log.Infof("Read brand : %s returned %+v\n", uuid, results)
	if err != nil {
		return Brand{}, false, err
	}
	if len(results) == 0 {
		return Brand{}, false, nil
	}
	return results[0].Brand, true, nil
}

func (s service) Write(thing interface{}) error {
	brand := thing.(Brand)
	brandProps := map[string]string{
		"uuid":           brand.UUID,
		"prefLabel":      brand.PrefLabel,
		"strapline":      brand.Strapline,
		"descriptionXML": brand.DescriptionXML,
		"description":    brand.Description,
		"imageUrl":       brand.ImageURL,
	}
	stmt := `
                OPTIONAL MATCH (:Brand {uuid:{uuid}})-[r:HAS_PARENT]->(:Brand)
                DELETE r
                MERGE (n:Thing {uuid: {uuid}})
                SET n:Brand
                SET n:Concept
                SET n={props}
                `
	params := neoism.Props{
		"uuid":  brand.UUID,
		"props": brandProps,
	}
	parentUUID := brand.ParentUUID
	if parentUUID != "" {
		stmt += `
                        MERGE (p:Thing {uuid:{parentUUID}})
                        MERGE (n)-[:HAS_PARENT]->(p)
                        `
		params["parentUUID"] = brand.ParentUUID
	}

	deleteIdentifiers := &neoism.CypherQuery{
		Statement: `MATCH (t:Thing {uuid:{uuid}})
                                OPTIONAL MATCH (i:Identifier)-[ir:IDENTIFIES]->(t)
                                DELETE ir, i`,
		Parameters: map[string]interface{}{
			"uuid": brand.UUID,
		},
	}

	writeQuery := &neoism.CypherQuery{
		Statement:  stmt,
		Parameters: params,
	}

	queries := []*neoism.CypherQuery{deleteIdentifiers, writeQuery}

	for _, identifier := range brand.Identifiers {
		queries = append(queries, identifierMerge(identifier, brand.UUID))
	}
	return s.cypherRunner.CypherBatch(queries)

}

const (
	fsAuthority  = "http://api.ft.com/system/FACTSET-PPL"
	tmeAuthority = "http://api.ft.com/system/FT-TME"
)

var identifierLabels = map[string]string{
	fsAuthority:  "FactsetIdentifier",
	tmeAuthority: "TMEIdentifier",
}

func identifierMerge(identifier identifier, uuid string) *neoism.CypherQuery {
	statementTemplate := fmt.Sprintf(`MERGE (o:Thing {uuid:{uuid}})
                                          MERGE (i:Identifier {value:{value} , authority:{authority}})
                                          MERGE (o)<-[:IDENTIFIES]-(i)
                                          set i:%s`, identifierLabels[identifier.Authority])

	query := &neoism.CypherQuery{
		Statement: statementTemplate,
		Parameters: map[string]interface{}{
			"uuid":      uuid,
			"value":     identifier.IdentifierValue,
			"authority": identifier.Authority,
		},
	}
	return query
}

func (s service) Delete(uuid string) (bool, error) {
	deleteIdentifiers := &neoism.CypherQuery{
		Statement: `MATCH (t:Thing {uuid:{uuid}})
                                OPTIONAL MATCH (i:Identifier)-[ir:IDENTIFIES]->(t)
                                WITH i, count(ir) as c, ir, t
                                WHERE c = 1
                                DELETE ir, i
                                `,
		Parameters: map[string]interface{}{
			"uuid": uuid,
		},
	}

	deleteOwnedRelationships := &neoism.CypherQuery{
		Statement: `
                        MATCH (n:Thing {uuid: {uuid}})-[p:HAS_PARENT]->(t:Thing)
                        DELETE p
                `,
		Parameters: neoism.Props{
			"uuid": uuid,
		},
		IncludeStats: true,
	}

	clearNode := &neoism.CypherQuery{
		Statement: `
			MATCH (n:Thing {uuid: {uuid}})
			REMOVE n:Brand
			SET n={props}
		`,
		Parameters: neoism.Props{
			"uuid": uuid,
			"props": neoism.Props{
				"uuid": uuid,
			},
		},
		IncludeStats: true,
	}

	removeNodeIfUnused := &neoism.CypherQuery{
		Statement: `
			MATCH (p:Thing {uuid: {uuid}})
			OPTIONAL MATCH (p)-[a]-(x)
			WITH p, count(a) AS relCount
			WHERE relCount = 0
			DELETE p
		`,
		Parameters: neoism.Props{
			"uuid": uuid,
		},
	}

	err := s.cypherRunner.CypherBatch([]*neoism.CypherQuery{deleteIdentifiers, deleteOwnedRelationships, clearNode, removeNodeIfUnused})

	s1, err := clearNode.Stats()
	if err != nil {
		return false, err
	}

	var deleted bool
	if s1.ContainsUpdates && s1.LabelsRemoved > 0 {
		deleted = true
	}

	return deleted, err
}

func (s service) DecodeJSON(dec *json.Decoder) (interface{}, string, error) {
	brand := Brand{}
	err := dec.Decode(&brand)
	return brand, brand.UUID, err
}

func (s service) Check() error {
	return neoutils.Check(s.cypherRunner)
}

func (s service) Count() (int, error) {

	results := []struct {
		Count int `json:"c"`
	}{}

	query := &neoism.CypherQuery{
		Statement: `MATCH (n:Brand) return count(n) as c`,
		Result:    &results,
	}

	err := s.cypherRunner.CypherBatch([]*neoism.CypherQuery{query})

	if err != nil {
		return 0, err
	}

	return results[0].Count, nil
}
