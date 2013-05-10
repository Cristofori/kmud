package testutils

func NewTestWriter(writeString *string) *TestWriter {
	var writer TestWriter
	writer.wrote = writeString
	return &writer
}

type TestWriter struct {
	wrote *string
}

func (self TestWriter) Write(p []byte) (n int, err error) {
	if self.wrote != nil {
		*self.wrote = string(p)
	}

	return len(p), nil
}

type TestReader struct {
	ToRead string
}

func (self TestReader) Read(p []byte) (n int, err error) {
	for i := 0; i < len(self.ToRead); i++ {
		p[i] = self.ToRead[i]
	}

	p[len(self.ToRead)] = '\n'

	return len(self.ToRead) + 1, nil
}

type TestReadWriter struct {
	TestReader
	TestWriter
}

// vim: nocindent
