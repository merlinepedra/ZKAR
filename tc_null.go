package zkar

type TCNull struct {
	// nothing
}

func (n *TCNull) ToBytes() []byte {
	return []byte{JAVA_TC_NULL}
}

func (n *TCNull) ToString() string {
	var b = NewPrinter()
	b.Printf("TC_NULL - %s", Hexify(JAVA_TC_NULL))
	return b.String()
}

func readTCNull(stream *ObjectStream) *TCNull {
	_, _ = stream.ReadN(1)
	return new(TCNull)
}
