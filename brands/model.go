package brands

// Brand structure used by API
type Brand struct {
	UUID           string `json:"uuid"`
	ParentUUID     string `json:"parentUUID"` // This becomes a relationship HAS_PARENT
	PrefLabel      string `json:"prefLabel"`
	Strapline      string `json:"strapline"`
	DescriptionXML string `json:"descriptionXML"`
	Description    string `json:"description"` // TODO this this is desirable for people who want to process pure text
	ImageURL       string `json:"imageUrl"`
}

var propsToRelationship = map[string]string{
	"parentUUID": "HAS_PARENT",
}
