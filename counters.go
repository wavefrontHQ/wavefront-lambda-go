package wflambda

type counter struct {
	val float64
}

func (c *counter) Increment(value int64) {
	c.val += float64(value)
}
