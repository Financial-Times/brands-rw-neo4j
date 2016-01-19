package brands

import (
	"encoding/json"

	"github.com/Financial-Times/neo-utils-go"
	"github.com/jmcvetta/neoism"
)

//Service maintains info about runners and index managers
type Service struct {
	cypherRunner neoutils.CypherRunner
	indexManager neoutils.IndexManager
}

// NewCypherBrandService provides functions for create, update, delete operations on people in Neo4j,
// plus other utility functions needed for a service
func NewCypherBrandService(cypherRunner neoutils.CypherRunner, indexManager neoutils.IndexManager) service {
	return Service{cypherRunner, indexManager}
}

func (s service) Initialise() error {
	return neoutils.EnsureConstraints(s.indexManager, map[string]string{
		"Thing":   "uuid",
		"Concept": "uuid",
		"Brand":   "uuid",
	})
}

func (s service) Read(uuid string) (interface{}, bool, error) {
	results := []struct {
		Brand
	}{}
	query := &neoism.CypherQuery{
		Statement: `
                MATCH (n:Brand {uuid:{uuid}}) RETURN n.uuid AS uuid,
                        n.parentUUID AS parentUUID, n.prefLabel AS prefLabel,
                        n.strapline AS strapLine, n.descriptionXML AS descriptionXML,
                        n.description AS description, n.imageUrl AS imageUrl
                        `,
		Parameters: map[string]interface{}{
			"uuid": uuid,
		},
		Result: &results,
	}
	err := s.cypherRunner.CypherBatch([]*neoism.CypherQuery{query})
	if err != nil {
		return Brand{}, false, err
	}
	if len(results) == 0 {
		return Brand{}, false, nil
	}
	return results[0], true, nil
}

func (s service) Write(thing interface{}) error {
	if parentUUID, exists := thing["parentUUID"]; exists {
		delete(thing, thing["parentUUID"])
	}
	query := &neoism.CypherQuery{
		Statement: `
                        MERGE (n:Thing {uuid: {uuid}})
                        SET n :Brand
                        SET n={allprops}
                        `,
		Parameters: map[string]interface{}{
			"uuid":     p.UUID,
			"allprops": thing,
		},
	}

	return s.cypherRunner.CypherBatch([]*neoism.CypherQuery{query})

}

func (s service) Delete(uuid string) (bool, error) {
	clearNode := &neoism.CypherQuery{
		Statement: `
			MATCH (n:Thing {uuid: {uuid}})
			REMOVE n:Brand
			SET n={props}
		`,
		Parameters: map[string]interface{}{
			"uuid": uuid,
			"props": map[string]interface{}{
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
		Parameters: map[string]interface{}{
			"uuid": uuid,
		},
	}

	err := s.cypherRunner.CypherBatch([]*neoism.CypherQuery{clearNode, removeNodeIfUnused})

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
	p := person{}
	err := dec.Decode(&p)
	return p, p.UUID, err
}

func (s service) Check() error {
	return neoutils.Check(s.cypherRunner)
}

func (s service) Count() (int, error) {

	results := []struct {
		Count int `json:"c"`
	}{}

	query := &neoism.CypherQuery{
		Statement: `MATCH (n:Person) return count(n) as c`,
		Result:    &results,
	}

	err := s.cypherRunner.CypherBatch([]*neoism.CypherQuery{query})

	if err != nil {
		return 0, err
	}

	return results[0].Count, nil
}

const (
	fsAuthority = "http://api.ft.com/system/FACTSET-PPL"
)
