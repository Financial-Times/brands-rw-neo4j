package brands

// Brand structure used by API
type Brand struct {
	UUID           string `json:"uuid"`
	PrefLabel      string `json:"prefLabel"`
	ParentUUID     string `json:"parentUUID"`
	Strapline      string `json:"strapline"`
	DescriptionXML string `json:"descriptionXML"`
	Description    string `json:"description"` // TODO this this is desirable for people who want to process pure text
	ImageURL       string `json:"_imageUrl"`   // TODO this is a temporary thing - needs to be integrated into images properly
}
