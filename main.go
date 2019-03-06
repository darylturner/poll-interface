package main

import (
	"flag"
	"fmt"
	"github.com/alouca/gosnmp"
	"log"
	"time"
)

func continuousPoll(s *gosnmp.GoSNMP, index *string, host *string) {
	ifnameoid := "1.3.6.1.2.1.31.1.1.1.1." + *index
	oids := []string{"1.3.6.1.2.1.31.1.1.1.6.", "1.3.6.1.2.1.31.1.1.1.10."}
	for i, _ := range oids {
		oids[i] = oids[i] + *index
	}

	ret, err := s.Get(ifnameoid)
	if err != nil {
		log.Fatal("interface not found")
	}
	ifname := ret.Variables[0].Value.(string)

	var iprev, oprev uint64
	iprev = 0
	oprev = 0

	ticker := time.NewTicker(time.Second * 10)
	for _ = range ticker.C {
		ret, err := s.GetMulti(oids)
		if err != nil {
			log.Fatal(err)
		}

		iraw := ret.Variables[0].Value.(uint64)
		oraw := ret.Variables[1].Value.(uint64)

		var idelta, odelta float64
		if iprev != 0 {
			idelta = float64(iraw - iprev)
			odelta = float64(oraw - oprev)

			fmt.Printf(
				"%v %v - in: %.3f mbps / out: %.3f mbps\n",
				*host, ifname,
				idelta*8/10/1e6,
				odelta*8/10/1e6,
			)
		}

		iprev = iraw
		oprev = oraw
	}
}

func discoverIndexes(s *gosnmp.GoSNMP) {
	ret, err := s.Walk("1.3.6.1.2.1.31.1.1.1.1")
	if err != nil {
		log.Fatal(err)
	}

	for _, result := range ret {
		fmt.Println(result.Name, result.Value.(string))
	}
}

func main() {
	host := flag.String("h", "", "host to poll")
	community := flag.String("c", "", "snmp community")
	index := flag.String("i", "", "interface index")
	discover := flag.Bool("d", false, "print interface discovery and exit")
	flag.Parse()

	if *index == "" && *discover == false {
		log.Fatal("please supply interface index number")
	}
	if *community == "" {
		log.Fatal("community can't be null")
	}
	if *host == "" {
		log.Fatal("community can't be null")
	}

	s, err := gosnmp.NewGoSNMP(*host, "network_services", gosnmp.Version2c, 5)
	if err != nil {
		log.Fatal(err)
	}

	if *discover {
		discoverIndexes(s)
	} else {
		continuousPoll(s, index, host)
	}
}
