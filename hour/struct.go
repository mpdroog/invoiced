package hour

type Hour struct {
	Name  string
	Lines []struct {
		Start       string
		Stop        string
		Hours       float64
		Description string
	}
}
