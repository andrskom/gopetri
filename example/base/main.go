package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

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
	pool := gopetri.NewPool(10, time.Second)
	if err := pool.Init(cfg); err != nil {
		log.Fatal(err.Error())
	}

	net, err := pool.Get()
	if err != nil {
		log.Fatal(err.Error())
	}
	defer net.Close()
	consumer.SetComp(net)
	net.SetConsumer(consumer)
	if err := net.StartPlace(); err != nil {
		log.Fatal(err.Error())
	}

	if err := net.SetPlace("branch2Place1"); err != nil {
		log.Fatal(err.Error())
	}
	if err := net.SetPlace("branch1Place1"); err != nil {
		log.Fatal(err.Error())
	}
	if err := net.SetPlace("branch1Place2"); err != nil {
		log.Fatal(err.Error())
	}
	if err := net.SetPlace("branchMergePlace1"); err != nil {
		log.Fatal(err.Error())
	}
	if err := net.SetPlace("placeFinish"); err != nil {
		log.Fatal(err.Error())
	}
	log.Println(net.IsFinished())

	log.Println("-----------")
	graphViz, err := net.AsGraphvizDotLang("example_v1", true)
	if err != nil {
		log.Fatal(err.Error())
	}
	fmt.Print(graphViz)
}
