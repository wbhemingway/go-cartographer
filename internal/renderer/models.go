package renderer

type Tile struct {
	X int `json:"x"`
	Y int `json:"y"`
	Terrain string `json:"terrain"`
	Creature string `json:"creature"`
	Structure string `json:"structure"`
}

type World struct {
	Width int `json:"width"`
	Height int `json:"height"`
	Tiles []Tile `json:"tiles"`
}
