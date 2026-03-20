package cartographer

type Tile struct {
	X         int    `json:"x"`
	Y         int    `json:"y"`
	Terrain   string `json:"terrain"`
	Creature  string `json:"creature,omitempty"`
	Structure string `json:"structure,omitempty"`
}

type WorldRequest struct {
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Tiles  []Tile `json:"tiles"`
}

type MapResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	URL    string `json:"url,omitempty"`
}