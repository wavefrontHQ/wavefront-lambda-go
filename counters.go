package wflambda

// counter is a struct to count values
type counter struct {
	val float64
}

// Increment updates the value of a counter with the given value
func (c *counter) Increment(value int64) {
	c.val += float64(value)
}
