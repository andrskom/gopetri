package gopetri

// Factory fro net with prepared buffer.
type Factory struct {
	cfg        Cfg
	preparedCh chan Net
	consumer   Consumer
}

// NewFactory init factory.
func NewFactory(cfg Cfg, consumer Consumer, numberOfPrepared int) *Factory {
	return &Factory{
		cfg:        cfg,
		consumer:   consumer,
		preparedCh: make(chan Net, numberOfPrepared),
	}
}

// Run async preparing of net.
func (f *Factory) Run() error {
	for {
		c, err := BuildFromCfg(f.cfg)
		if err != nil {
			return err
		}
		c.SetConsumer(f.consumer)
		f.preparedCh <- *c
	}
}

// Get return prepared petri net.
func (f *Factory) Get() *Net {
	c := <-f.preparedCh
	return &c
}
