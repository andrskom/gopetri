package util

import (
	"log"

	"github.com/andrskom/gopetri"
)

type LogConsumer struct {
	comp *gopetri.Component
}

func (l *LogConsumer) SetComp(comp *gopetri.Component) {
	l.comp = comp
}

func (l *LogConsumer) BeforePlace(placeID string) error {
	return nil
}

func (l *LogConsumer) AfterPlace(placeID string) {
	log.Printf(`-> Place
   %s
   %#v
`, placeID, l.comp.GetState())
}

func (l *LogConsumer) CanTransit(transitionID string) bool {
	return true
}

func (l *LogConsumer) BeforeTransit(transitionID string) error {
	return nil
}

func (l *LogConsumer) AfterTransit(transitionID string) {
	log.Printf(`+> Transit
   %s
   %#v
`, transitionID, l.comp.GetState())
}
