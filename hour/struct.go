package hour

type Hour struct {
	Name  string `validate:"slug,nonzero"`
	Lines []struct {
		Day         string `validate:"date,nonzero"`
		Start       string `validate:"time,nonzero"`
		Stop        string `validate:"time,nonzero"`
		Hours       float64 `validate:"uint,nonzero"`
		Description string
	}
}
