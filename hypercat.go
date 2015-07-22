package hypercat

import (
	"encoding/json"
	"errors"
)

const (
	// HyperCatVersion is the version of HyperCat this library currently supports
	HyperCatVersion = "2.0"

	// HyperCatMediaType is the default mime type of HyperCat resources
	HyperCatMediaType = "application/vnd.hypercat.catalogue+json"

	// DescriptionRel is the URI for the hasDescription relationship
	DescriptionRel = "urn:X-hypercat:rels:hasDescription:en"

	// ContentTypeRel is the URI for the isContentType relationship
	ContentTypeRel = "urn:X-hypercat:rels:isContentType"

	// HomepageRel is the URI for hasHomepage relationship
	HomepageRel = "urn:X-hypercat:rels:hasHomepage"

	// ContainsContentTypeRel is the URI for the containsContentType relationship
	ContainsContentTypeRel = "urn:X-hypercat:rels:containsContentType"

	// SupportsSearchRel is the URI for the supportsSearch relationship
	SupportsSearchRel = "urn:X-hypercat:rels:supportsSearch"

	// SimpleSearchVal is the required value for catalogues that support HyperCat simple search.
	SimpleSearchVal = "urn:X­hypercat:search:simple"

	// GeoBoundSearchVal is the required value for catalogues that support geographic bounding box search
	GeoBoundSearchVal = "urn:X-hypercat:search:geobound"

	// LexicographicSearchVal is the required value for catalogues that support lexicographic searching
	LexicographicSearchVal = "urn:X­hypercat:search:lexrange"

	// MultiSearchVal is the required value for catalogues that support multi-search
	MultiSearchVal = "urn:X-hypercat:search:multi"

	// SubstringSearchVal is the required value for catalogues that support substring search
	SubstringSearchVal = "urn:X-hypercat:search:substring"
)

/*
 * HyperCat is the representation of the HyperCat catalogue object, which is
 * the parent element of each catalogue instance.
 */
type HyperCat struct {
	Items       Items    `json:"items"`
	Metadata    Metadata `json:"item-metadata"`
	Description string   `json:"-"` // HyperCat spec is fuzzy about whether there can be more than one description. We assume not.
}

/*
 * NewHyperCat is a constructor function that creates and returns a HyperCat
 * instance.
 */
func NewHyperCat(description string) *HyperCat {
	return &HyperCat{
		Description: description,
		Metadata:    Metadata{},
	}
}

/*
 * Parse is a function that parses a HyperCat catalogue string, and builds an
 * in memory HyperCat instance.
 */
func Parse(str string) (*HyperCat, error) {
	cat := HyperCat{}
	err := json.Unmarshal([]byte(str), &cat)

	if err != nil {
		return nil, err
	}

	return &cat, nil
}

/*
 * AddRel is a function for adding a Rel object to a catalogue. This may result
 * in duplicated Rel keys as this is permitted by the HyperCat spec.
 * TODO: this code is duplicated in item
 */
func (h *HyperCat) AddRel(rel, val string) {
	h.Metadata = append(h.Metadata, Rel{Rel: rel, Val: val})
}

/*
 * ReplaceRel is a function that attempts to replace the value of a specific
 * Rel object if it is attached to this Catalogue. If the Rel key isn't found
 * this will have no effect.
 */
func (h *HyperCat) ReplaceRel(rel, val string) {
	for i, relationship := range h.Metadata {
		if relationship.Rel == rel {
			h.Metadata[i] = Rel{Rel: rel, Val: val}
		}
	}
}

/*
 * AddItem is a function for adding an Item to a catalogue. Returns an error if
 * we try to add an Item whose href is already defined within the catalogue.
 */
func (h *HyperCat) AddItem(item *Item) error {
	for _, i := range h.Items {
		if item.Href == i.Href {
			err := errors.New(`An item with href: "` + item.Href + `" is a already defined within the catalogue`)
			return err
		}
	}

	h.Items = append(h.Items, *item)

	return nil
}

/*
 * ReplaceItem is a function for replacing an item within a catalogue. Returns an error
 * if we try to replace an Item that isn't defined within the catalogue.
 */
func (h *HyperCat) ReplaceItem(newItem *Item) error {
	for index, oldItem := range h.Items {
		if newItem.Href == oldItem.Href {
			h.Items[index] = *newItem
			return nil
		}
	}

	err := errors.New(`An item with href: "` + newItem.Href + `" was not found within the catalogue`)
	return err
}

/*
 * MarshalJSON returns the JSON encoding of a HyperCat. This function is the
 * implementation of the Marshaler interface.
 */
func (h *HyperCat) MarshalJSON() ([]byte, error) {
	metadata := h.Metadata

	if h.Description != "" {
		metadata = append(metadata, Rel{Rel: DescriptionRel, Val: h.Description})
	}

	return json.Marshal(struct {
		Items    []Item   `json:"items"`
		Metadata Metadata `json:"item-metadata"`
	}{
		Items:    h.Items,
		Metadata: metadata,
	})
}

/*
 * UnmarshalJSON is the required function for structs that implement the
 * Unmarshaler interface.
 */
func (h *HyperCat) UnmarshalJSON(b []byte) error {
	type tempCat struct {
		Items    Items    `json:"items"`
		Metadata Metadata `json:"item-metadata"`
	}

	t := tempCat{}

	err := json.Unmarshal(b, &t)

	if err != nil {
		return err
	}

	for _, rel := range t.Metadata {
		if rel.Rel == DescriptionRel {
			h.Description = rel.Val
		} else {
			h.Metadata = append(h.Metadata, rel)
		}
	}

	if h.Description == "" {
		err := errors.New(`"` + DescriptionRel + `" is a mandatory metadata element`)
		return err
	}

	return nil
}

/*
 * Rels returns a slice containing all the Rel values of catalogue's metadata.
 */
func (h *HyperCat) Rels() []string {
	rels := make([]string, len(h.Metadata))

	for i, rel := range h.Metadata {
		rels[i] = rel.Rel
	}

	return rels
}

/*
 * Vals returns a slice of all values that match the given rel value.
 */
func (h *HyperCat) Vals(key string) []string {
	vals := []string{}

	for _, rel := range h.Metadata {
		if rel.Rel == key {
			vals = append(vals, rel.Val)
		}
	}

	return vals
}
