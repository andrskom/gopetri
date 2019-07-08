package gopetri

type Factory struct {
	cfg        Cfg
	preparedCh chan Component
	consumer   Consumer
}

func NewFactory(cfg Cfg, consumer Consumer, numberOfPrepared int) *Factory {
	return &Factory{
		cfg:        cfg,
		consumer:   consumer,
		preparedCh: make(chan Component, numberOfPrepared),
	}
}

func (f *Factory) Run() error {
	for {
		c, err := BuildFromCfg(f.cfg, f.consumer)
		if err != nil {
			return err
		}
		f.preparedCh <- *c
	}
}

func (f *Factory) Get() *Component {
	c := <-f.preparedCh
	return &c
}
