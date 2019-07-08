package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/andrskom/gopetri"
	"github.com/andrskom/gopetri/example/jsonsrc"
	"github.com/andrskom/gopetri/util"
)

func main() {
	cfgBytes, err := jsonsrc.GetByID("example_v1")
	if err != nil {
		log.Fatal(err.Error())
	}

	var cfg gopetri.Cfg
	if err := json.Unmarshal(cfgBytes, &cfg); err != nil {
		log.Fatal(err.Error())
	}

	consumer := &util.LogConsumer{}
	fac := gopetri.NewFactory(cfg, consumer, 3)
	go fac.Run()

	comp := fac.Get()
	consumer.SetComp(comp)
	if err := comp.StartPlace(); err != nil {
		log.Fatal(err.Error())
	}
	finished, err := comp.SetPlace("branch2Place1")
	if err != nil {
		log.Fatal(err.Error())
	}
	finished, err = comp.SetPlace("branch1Place1")
	if err != nil {
		log.Fatal(err.Error())
	}
	finished, err = comp.SetPlace("branch1Place2")
	if err != nil {
		log.Fatal(err.Error())
	}
	finished, err = comp.SetPlace("branchMergePlace1")
	if err != nil {
		log.Fatal(err.Error())
	}
	finished, err = comp.SetPlace("placeFinish")
	if err != nil {
		log.Fatal(err.Error())
	}
	log.Println(finished)

	log.Println("-----------")
	graphViz, err := comp.AsGraphvizDotLang("example_v1", true)
	if err != nil {
		log.Fatal(err.Error())
	}
	fmt.Print(graphViz)
}
