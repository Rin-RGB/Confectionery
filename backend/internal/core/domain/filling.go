package domain

type Filling struct {
	ID          string
	Name        string
	Description string
	Price       int64
	ImageURL    string
	IsActive    bool
}

type FillingPatch struct {
	Name        *string
	Description *string
	Price       *int64
	ImageURL    *string
	IsActive    *bool
}
