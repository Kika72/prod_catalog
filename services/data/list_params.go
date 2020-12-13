package data

type SortingDirection int

const (
	Descending SortingDirection = -1
	Ascending  SortingDirection = 1
)

type OrderParam struct {
	FieldName string
	Direction SortingDirection
}
