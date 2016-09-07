package brands

import (
	"encoding/json"
	"fmt"
	"github.com/Financial-Times/neo-utils-go/neoutils"
	"github.com/jmcvetta/neoism"
)

//service maintains info about runners and index managers
type service struct {
	conn neoutils.NeoConnection
}

// NewCypherBrandsService provides functions for create, update, delete operations on brands in Neo4j,
// plus other utility functions needed for a service
func NewCypherBrandsService(cypherRunner neoutils.NeoConnection) service {
	return service{cypherRunner}
}

//Initialise the driver
func (s service) Initialise() error {

	err := s.conn.EnsureIndexes(map[string]string{
		"Identifier": "value",
	})

	if err != nil {
		return err
	}

	return s.conn.EnsureConstraints(map[string]string{
		"Thing":         "uuid",
		"Concept":       "uuid",
		"Brand":         "uuid",
		"TMEIdentifier": "value",
		"UPPIdentifier": "value"})
}

func (s service) Read(uuid string) (interface{}, bool, error) {
	results := []struct {
		Brand
	}{}
	query := &neoism.CypherQuery{
		Statement: `
                        MATCH (n:Brand {uuid:{uuid}})
                        OPTIONAL MATCH (n)-[:HAS_PARENT]->(p:Thing)
                        OPTIONAL MATCH (upp:UPPIdentifier)-[:IDENTIFIES]->(n)
			OPTIONAL MATCH (tme:TMEIdentifier)-[:IDENTIFIES]->(n)
                        RETURN n.uuid AS uuid, n.prefLabel AS prefLabel,
                                n.strapline AS strapline, p.uuid as parentUUID,
                                n.descriptionXML AS descriptionXML,
                                n.description AS description, n.imageUrl AS _imageUrl,
                                {uuids:collect(distinct upp.value), TME:collect(distinct tme.value)} as alternativeIdentifiers,
                                labels(n) as types
                                `,
		Parameters: map[string]interface{}{
			"uuid": uuid,
		},
		Result: &results,
	}
	err := s.conn.CypherBatch([]*neoism.CypherQuery{query})
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

	deleteParentRelationship := &neoism.CypherQuery{
		Statement: `
                        MATCH (:Thing {uuid:{uuid}})-[r:HAS_PARENT]->(:Thing)
                        DELETE r`,
		Parameters: neoism.Props{
			"uuid": brand.UUID,
		},
	}

	deleteIdentifiers := &neoism.CypherQuery{
		Statement: `
                        MATCH (t:Thing {uuid:{uuid}})
                        OPTIONAL MATCH (i:Identifier)-[ir:IDENTIFIES]->(t)
                        DELETE ir, i`,
		Parameters: neoism.Props{
			"uuid": brand.UUID,
		},
	}

	writeBrand := &neoism.CypherQuery{
		Statement: `
                        MERGE (n:Thing {uuid: {uuid}})
                        SET n:Brand
                        SET n:Concept
			SET n:Classification
                        SET n={props}`,
		Parameters: neoism.Props{
			"uuid":  brand.UUID,
			"props": brandProps,
		},
	}
	queries := []*neoism.CypherQuery{deleteParentRelationship, deleteIdentifiers, writeBrand}

	if len(brand.ParentUUID) > 0 {
		fmt.Printf("**HAS PARENT %s", brand.ParentUUID)
		writeParent := &neoism.CypherQuery{
			Statement: `
                                MATCH (t:Thing {uuid:{uuid}})
                                MERGE (p:Thing {uuid:{parentUUID}})
                                MERGE (t)-[:HAS_PARENT]->(p)`,
			Parameters: neoism.Props{
				"parentUUID": brand.ParentUUID,
				"uuid":       brand.UUID,
			},
		}
		queries = append(queries, writeParent)
	}

	//ADD all the IDENTIFIER nodes and IDENTIFIES relationships
	for _, alternativeUUID := range brand.AlternativeIdentifiers.TME {
		alternativeIdentifierQuery := createNewIdentifierQuery(brand.UUID, tmeIdentifierLabel, alternativeUUID)
		queries = append(queries, alternativeIdentifierQuery)
	}

	for _, alternativeUUID := range brand.AlternativeIdentifiers.UUIDS {
		alternativeIdentifierQuery := createNewIdentifierQuery(brand.UUID, uppIdentifierLabel, alternativeUUID)
		queries = append(queries, alternativeIdentifierQuery)
	}

	return s.conn.CypherBatch(queries)
}

func createNewIdentifierQuery(uuid string, identifierLabel string, identifierValue string) *neoism.CypherQuery {
	statementTemplate := fmt.Sprintf(`MERGE (t:Thing {uuid:{uuid}})
					CREATE (i:Identifier {value:{value}})
					MERGE (t)<-[:IDENTIFIES]-(i)
					set i : %s `, identifierLabel)
	query := &neoism.CypherQuery{
		Statement: statementTemplate,
		Parameters: map[string]interface{}{
			"uuid":  uuid,
			"value": identifierValue,
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
			REMOVE n:Brand:Concept:Classification
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

	err := s.conn.CypherBatch([]*neoism.CypherQuery{deleteIdentifiers, deleteOwnedRelationships, clearNode, removeNodeIfUnused})

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
	return neoutils.Check(s.conn)
}

func (s service) Count() (int, error) {

	results := []struct {
		Count int `json:"c"`
	}{}

	query := &neoism.CypherQuery{
		Statement: `MATCH (n:Brand) return count(n) as c`,
		Result:    &results,
	}

	err := s.conn.CypherBatch([]*neoism.CypherQuery{query})

	if err != nil {
		return 0, err
	}

	return results[0].Count, nil
}
