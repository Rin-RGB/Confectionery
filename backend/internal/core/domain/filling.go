package domain

type Filling struct {
	ID          string
	Name        string
	Description string
	Price       float64
	ImageName   string
	IsActive    bool
}

type FillingPatch struct {
	Name        *string
	Description *string
	Price       *float64
	ImageName   *string
	IsActive    *bool
}
