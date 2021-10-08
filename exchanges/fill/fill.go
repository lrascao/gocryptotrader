package fill

func (f *Fills) Setup(exchangeName string, fillsFeedEnabled bool, c chan interface{}) error {
	f.exchangeName = exchangeName
	f.dataHandler = c
	f.fillsFeedEnabled = fillsFeedEnabled

	return nil
}

func (f *Fills) Update(data ...Data) error {
	if f.fillsFeedEnabled {
		f.dataHandler <- data
	}

	return nil
}
