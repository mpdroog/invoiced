package hour

type Hour struct {
	Name  string
	Lines []struct {
		Day         string
		Start       string
		Stop        string
		Hours       float64
		Description string
	}
}
