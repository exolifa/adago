package main

import (
	"fmt"

	"adago.net/adago/dbproc"
	"adago.net/adago/routing"
)

func main() {

	routeur := routing.SetRoutes()
	myport := ":" + dbproc.GetConfig("Tcpport")
	routeur.Run(myport)
	fmt.Printf("Running gin route on port 8999")
	/*
		mycarlist := dbproc.Carslist(Myconn, 0)
		for i := 0; i < len(mycarlist.Records); i++ {
			fmt.Printf("|%v \t\t|%v\t\t|%v\t\t|\n", mycarlist.Records[i].CD.Vendor, mycarlist.Records[i].CD.Model, mycarlist.Records[i].CD.Color)
		}
	*/
}
