package domain

type Detection struct {
	Desc  string `json:"desc"`
	Box   []int  `json:"box"`
	Box2D []int  `json:"box_2d"`
}
