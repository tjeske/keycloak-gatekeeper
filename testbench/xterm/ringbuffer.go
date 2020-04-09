package main

type Ringbuf struct {
	Input  chan<- interface{}
	Output <-chan interface{}
}

// Creates a new ring buffer of the specified size
// NB: `nil' values are ignored/filtered out
func NewRingBuffer(size int) (result *Ringbuf) {
	input := make(chan interface{}, 0)
	output := make(chan interface{}, size)
	result = &Ringbuf{
		Input:  input,
		Output: output,
	}
	go maintain(input, output)
	return
}

// Makes sure that all writes to Input are processed before continuing
func (rb Ringbuf) Flush() {
	rb.Input <- nil
}

func (rb Ringbuf) Write(p []byte) (n int, err error) {
	rb.Input <- p
	return len(p), nil
}

// maintains the input-to-output transfer
func maintain(input <-chan interface{}, output chan interface{}) {
	for n := range input {
		for n != nil {
			select {
			case output <- n:
				n = nil
			default:
				// can't write yet?
				select {
				case <-output:
					// try to remove element from output buffer
				default:
					// output buffer is empty... that reader was fast!
				}
			}
		}
	}
	close(output)
}
