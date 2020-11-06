//Package dbproc contains all the procedure used to connect to ADABAS and make basic CRUD operation.
package dbproc

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/SoftwareAG/adabas-go-api/adabas"
	"github.com/SoftwareAG/adabas-go-api/adatypes"
)

// Carinfo contains data on car
type Carinfo struct {
	Isn    uint64
	Vendor string
	Model  string
	Color  string
}

// carrec contains data on car
type carrec struct {
	Vendor string
	Model  string
	Color  string
}

// CarList is the structure to get the list of cars
type CarList struct {
	Vehicules []*Carinfo
	Message   string
}

// switchprop contains a database switching proposal
type switchprop struct {
	Num   int    `json:"Numero"`
	SWUrl string `json:"Url"`
	Fnr   int    `json:"Fnr"`
}

// switchList contains list of mnemonic mapping
var switchList map[string]*switchprop

// config is the structure for all globals parameters
type config struct {
	Dbtouse       string `json:"db_to_use"`
	Dburl         string
	ConnectString string
	Switcherconf  string `json:"switcher_config"`
	Mapsfile      string `json:"maps_file"`
	Templatesdir  string `json:"templates"`
	Tcpport       string `json:"port"`
	Fnr           uint32
}

// Conf contains parameters for all modules
var Conf config

// GetConfig allows other modules to get access to parameters in config file
func GetConfig(item string) string {
	switch item {
	case "dbtouse":
		return Conf.Dbtouse
	case "dburl":
		return Conf.Dburl
	case "connectString":
		return Conf.ConnectString
	case "Mapsfile":
		return Conf.Mapsfile
	case "Templatesdir":
		return Conf.Templatesdir
	case "Tcpport":
		return Conf.Tcpport
	case "Fnr":
		return fmt.Sprintf("%v", Conf.Fnr)
	}
	return "unknown request"
}

// init load the json file for switching databases
func init() {
	conffyle := os.Args[1]
	fic, ficerr := ioutil.ReadFile(conffyle)
	if ficerr != nil {
		fmt.Printf("error ioutil : %v \n", ficerr)
	}
	_ = json.Unmarshal([]byte(fic), &Conf)
	switcherfyle := Conf.Switcherconf
	fichier, _ := ioutil.ReadFile(switcherfyle)
	_ = json.Unmarshal([]byte(fichier), &switchList)
	fmt.Printf("Init summary\n=================\n")
	fmt.Printf("Received parameter:%v\n", conffyle)
	fmt.Printf("Param file title :%v\n", string(fic))
	fmt.Printf("Switch config file :%v\n", switcherfyle)
	fmt.Printf("Switch file content:%v\n", string(fichier))
}

// result2struct transforms adabas response into array of struct to enable easy publishing in HTML templates
func result2struct(r *adabas.Response) []*Carinfo {
	var cl []*Carinfo
	if r.NrRecords() > 0 {
		for _, v := range r.Values {
			isn := v.Isn
			ven, _ := v.SearchValue("Vendor")
			mod, _ := v.SearchValue("Model")
			col, _ := v.SearchValue("Color")
			var voiture Carinfo
			voiture.Isn = uint64(isn)
			voiture.Vendor = ven.String()
			voiture.Model = mod.String()
			voiture.Color = col.String()
			cl = append(cl, &voiture)
		}
	}
	return cl
}

// adaswitch create a adabas connection string based on received value(s)
func adaswitch(i interface{}) {
	switch v := i.(type) {
	case map[string]string:
		mi := make(map[string]string)
		for k, w := range v {
			mi[k] = w
		}
		Conf.Dburl = mi["num"] + "(adatcp://" + mi["url"] + ")"
		Conf.ConnectString = mi["num"] + "(adatcp://" + mi["url"] + ")," + mi["fnr"]
		i64, _ := strconv.ParseInt(mi["fnr"], 10, 32)
		Conf.Fnr = uint32(i64)
	default:
		clef := fmt.Sprintf("%v", i)
		Conf.Dburl = fmt.Sprintf("%v(adatcp://%v)", switchList[clef].Num, switchList[clef].SWUrl)
		Conf.ConnectString = fmt.Sprintf("%v,%v", Conf.Dburl, switchList[clef].Fnr)
		Conf.Fnr = uint32(switchList[clef].Fnr)
	}
	fmt.Printf("configuring Dburl to %v\n", Conf.Dburl)
	fmt.Printf("configuring Dbconn to %v\n", Conf.ConnectString)
}

// loadMyJSONMap will load the maps to the DB
func loadMyJSONMap(Myconn *adabas.Connection) {
	fmt.Printf("loading map from %v into :%v\n", Conf.Mapsfile, Conf.Dburl)
	var dburl, _ = adabas.NewURL(Conf.Dburl)
	var maprepo = adabas.NewMapRepositoryWithURL(adabas.DatabaseURL{URL: *dburl, Fnr: adabas.Fnr(Conf.Fnr)})
	var monmap, maperr = adabas.LoadJSONMap(Conf.Mapsfile)
	if maperr != nil {
		fmt.Printf("load json error:%v\n", maperr)
	}
	monmap[0].Repository = &maprepo.DatabaseURL
	fmt.Println("ADDING MAP TO GLOBAL REPO")
	fmt.Println("\tADDING FNR 4 AS GLOBAL REPO")
	addmaperr := adabas.AddGlobalMapRepositoryReference(Conf.ConnectString)
	if addmaperr != nil {
		fmt.Printf("global map add error:%v\n", addmaperr)
	}
	fmt.Println("\tSTORE MAP TO REPO")
	storeerr := monmap[0].Store()
	if storeerr != nil {
		fmt.Printf("store map error=%v\n", storeerr)
	}
	fmt.Println("ADDING MAP TO CACHE")
	maprepo.AddMapToCache("DempMap", monmap[0])
}

// Adabasinit is used to initialize Adabas connection
func Adabasinit() *adabas.Connection {
	adaswitch(Conf.Dbtouse)
	// creating the connection
	fullconn := "acj;map;config=[" + Conf.ConnectString + "]"
	fmt.Printf("Opening connection to :%v\n", fullconn)
	adaConnect, err := adabas.NewConnection(fullconn)
	if err != nil {
		fmt.Printf("NewConnection() error=%v\n", err)
	}
	loadMyJSONMap(adaConnect)
	return adaConnect
}

// Carslist return the list of cars up to number given by limite - zero means all
func Carslist(Myconn *adabas.Connection, limite uint64) CarList {
	// creating Read Request with MAP
	readRequest, cerr := Myconn.CreateMapReadRequest(&Carinfo{})
	if cerr != nil {
		fmt.Printf("CreateMapReadRequest() error=%v\n", cerr)
	}
	// Assigning query's fields
	err := readRequest.QueryFields("Vendor,Model,Color")
	if err != nil {
		fmt.Printf("QueryFields() error=%v\n", err)
	}
	readRequest.Limit = limite
	// Performing the query ordered by NAME
	result, err := readRequest.ReadLogicalBy("Vendor")
	if err != nil {
		fmt.Printf("ReadLogicalWith() error=%v\n", err)
	}
	var carlist CarList
	//	carlist.Vehicules = result2struct(result)
	carlist.Vehicules = make([]*Carinfo, 0)
	for _, v := range result.Data {
		carlist.Vehicules = append(carlist.Vehicules, v.(*Carinfo))
	}
	return carlist
}

// CarsSearch return the list of cars for the specified vendors
func CarsSearch(Myconn *adabas.Connection, vv string, mv string, cv string, limite uint64) CarList {
	// creating Read Request with MAP
	readRequest, cerr := Myconn.CreateMapReadRequest(&Carinfo{})
	if cerr != nil {
		fmt.Printf("CreateMapReadRequest() error=%v\n", cerr)
	}
	// Assigning query's fields
	err := readRequest.QueryFields("Vendor,Model,Color")
	if err != nil {
		fmt.Printf("QueryFields() error=%v\n", err)
	}
	readRequest.Limit = limite
	// Performing the query ordered by NAME
	cherche := ""
	if vv != "" {
		cherche = cherche + "Vendor=" + vv
		if cv != "" || mv != "" {
			cherche = cherche + " AND "
		}
	}
	if mv != "" {
		cherche = cherche + "Model=" + mv
		if cv != "" {
			cherche = cherche + " AND "
		}
	}
	if cv != "" {
		cherche = cherche + "Color=" + cv
	}
	fmt.Printf("car search with :%v\n", cherche)
	result, err := readRequest.ReadLogicalWith(cherche)
	if err != nil {
		fmt.Printf("ReadLogicalWith() error=%v\n", err)
	}
	var carlist CarList
	carlist.Vehicules = make([]*Carinfo, 0)
	for _, v := range result.Data {
		carlist.Vehicules = append(carlist.Vehicules, v.(*Carinfo))
	}
	return carlist
}

//AddCar tries to create a car record in the file
func AddCar(Myconn *adabas.Connection, vendeur string, modele string, couleur string) string {
	// creating Store Request with MAP
	storeRequest, cerr := Myconn.CreateMapStoreRequest(&Carinfo{})
	if cerr != nil {
		return cerr.Error()
	}
	// Assigning query's fields
	err := storeRequest.StoreFields("Vendor,Model,Color")
	if err != nil {
		return err.Error()
	}
	enreg := &carrec{Vendor: vendeur, Model: modele, Color: couleur}
	sterr := storeRequest.StoreData(enreg)
	if sterr != nil {
		return sterr.Error()
	}
	transerr := storeRequest.EndTransaction()
	if transerr != nil {
		return transerr.Error()
	}
	return ""
}

// DelCar enables user to delete a car
func DelCar(Myconn *adabas.Connection, carisn uint64) string {
	deleteRequest, cerr := Myconn.CreateMapDeleteRequest("VehiclesMap")
	if cerr != nil {
		return cerr.Error()
	}

	sterr := deleteRequest.Delete(adatypes.Isn(carisn))
	if sterr != nil {
		return sterr.Error()
	}
	transerr := deleteRequest.EndTransaction()
	if transerr != nil {
		return transerr.Error()
	}
	return ""
}

//UpdateCar  enables update
func UpdateCar(Myconn *adabas.Connection, isn uint64, vendeur string, modele string, couleur string) string {
	// creating Read Request with MAP
	readRequest, cerr := Myconn.CreateMapReadRequest(&Carinfo{})
	if cerr != nil {
		return cerr.Error()
	}
	// Assigning query's fields
	readRequest.SetHoldRecords(adatypes.HoldWait)
	err := readRequest.QueryFields("Vendor,Model,Color")
	if err != nil {
		return err.Error()
	} // Performing the query ordered by NAME
	readresult, err := readRequest.ReadISN(adatypes.Isn(isn))
	if err != nil {
		return err.Error()
	}
	enreg := readresult.Values[0]
	fmt.Printf("enreg:%v\n", enreg)
	enreg.SetValue("Vendor", vendeur)
	enreg.SetValue("Model", modele)
	enreg.SetValue("Color", couleur)
	fmt.Printf("enreg modified:%v\n", enreg)
	//  the mode of creation of the storeRequest must be the same than the mode of creation of the previous ReadRequest
	storeRequest, cerr := Myconn.CreateMapStoreRequest(&Carinfo{})
	if cerr != nil {
		return cerr.Error()
	}
	storeRequest.StoreFields("Vendor,Model,Color")
	err = storeRequest.Update(enreg)
	if err != nil {
		return err.Error()
	}
	err = storeRequest.EndTransaction()
	if err != nil {
		return err.Error()
	}
	Myconn.Release()
	return ""
}
