package main

import (
	"fmt"

	"adago.net/adago/dbproc"
	"adago.net/adago/routing"
)

/*
* 	Working with ADABAS from a Golang program seems to be a challenge and usually it is anticipated to require some software
in between to translate the requests and the responses. This is one of the purpose of SQLGateway .
	This is indeed a reasonable approach to make ADABAS "transparent" to the Golang developper so he/she does not need
to learn how ADABAS works .
	The availability of a ADABAS-GO-API (likewise JAVA Client for ADABAS) opens a different perspective and allow to directly connect
the two worlds.
	With limited developer's skills and very restricted knowledges of ADABAS I decided to prepare a small, basic demo using
this API.
	The purpose of the demo is to have a small web Application ,developped in Go, and enabling basic CRUD opérations on the CARs file
in the demo database of ADABAS-CE (running in container)
	This demo uses Gin-gonic to handle the requests/responses.
	The heart of the solution is in the dbproc module where the ADABAS calls are handled.
	Structure of the project:
	- adago.go : is the main program to assemble all the modules
	- Routing     Module: handle the gin-gonic routing's instructions
	- Processors  Module: Contains the different handlers for the different routes
	- dbproc      Module: Handle all the interaction with ADABAS
	- templates : the html templates used
	- adagoparam.json : setup the different environmental parameteres . This file is passed as parameter to the adago program
	- vehicules.json  : this is the map used to hide the native ADABAS labels
	- switcherconf.json : The objective is to be able to execute the program against different DBs without recompile.
						  This file contains mnemonic and connection's instructions for the different databases
*/

func main() {
// create the http listener and start the router on the desired port
	routeur := routing.SetRoutes()
	myport := ":" + dbproc.GetConfig("Tcpport")
	routeur.Run(myport)
	fmt.Printf("Running gin route on port 8999")
}
