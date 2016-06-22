package brands

// Brand structure used by API
type Brand struct {
	UUID                   string                 `json:"uuid"`
	PrefLabel              string                 `json:"prefLabel"`
	Description            string                 `json:"description"`
	ParentUUID             string                 `json:"parentUUID,omitempty"`
	Strapline              string                 `json:"strapline"`
	DescriptionXML         string                 `json:"descriptionXML"`
	ImageURL               string                 `json:"_imageUrl"` // TODO this is a temporary thing - needs to be integrated into images properly
	AlternativeIdentifiers alternativeIdentifiers `json:"alternativeIdentifiers"`
	Types                  []string               `json:"types,omitempty"`
}

type alternativeIdentifiers struct {
	UUIDS []string `json:"uuids"`
	TME   []string `json:"TME,omitempty"`
}

const (
	uppIdentifierLabel = "UPPIdentifier"
	tmeIdentifierLabel = "TMEIdentifier"
)
