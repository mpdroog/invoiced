package hour

type Hour struct {
	Project  string `validate:"slug,nonzero"`
	Name     string `validate:"slug,nonzero"`
	Status   string `validate:"slug,nonzero"`
	Total    string `validate:"qty,nonzero"`
	HourRate float64 // cache

	Lines   []struct {
		Day         string  `validate:"date,nonzero"`
		Start       string  `validate:"time,nonzero"`
		Stop        string  `validate:"time,nonzero"`
		Hours       float64 `validate:"uint,nonzero"`
		Description string
	}
}
