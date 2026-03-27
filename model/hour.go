package model

type HourLine struct {
	Day         string  `validate:"date,nonzero"`
	Start       string  `validate:"time,nonzero"`
	Stop        string  `validate:"time,nonzero"`
	Hours       float64 `validate:"uint,nonzero"`
	Description string
}

type Hour struct {
	Project   string `validate:"slug,nonzero"`
	Name      string `validate:"slug,nonzero"`
	Status    string `validate:"slug,nonzero"`
	Total     string `validate:"qty,nonzero"`
	HourRate  float64 // cache
	Issuedate string  // YYYY-MM-DD, derived from filename if not set

	Lines []HourLine
}
