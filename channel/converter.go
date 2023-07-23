package channel

type Converter[IN, OUT any] func(IN) (OUT, error)

type Opt[IN, OUT any] func(IN, OUT) error

func Convert[IN, OUT any](srcC <-chan IN, convert Converter[IN, OUT], opts ...Opt[IN, OUT]) (<-chan OUT, chan error) {
	dstC, errC := make(chan OUT, cap(srcC)), make(chan error)
	go func() {
		defer close(errC)
		defer close(dstC)
		for {
			val, ok := <-srcC
			if !ok {
				return
			}
			cval, err := convert(val)
			if err != nil {
				errC <- err
			}
			for _, opt := range opts {
				if err := opt(val, cval); err != nil {
					errC <- err
				}
			}
			dstC <- cval
		}
	}()
	return dstC, errC
}
