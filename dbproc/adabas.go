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

// Carrec contains data on car
type Carrec struct {
	Vendor string
	Model  string
	Color  string
}

//CarRecU contain update record
type CarRecU struct {
	Isn    adatypes.Isn
	Vendor string
	Model  string
	Color  string
}

// CarList is the structure to get the list of cars
type CarList struct {
	Vehicules []Carinfo
}

// UpdCar is the structure to return result of create
type UpdCar struct {
	Opcode   bool
	Vehicule Carinfo
}

// Switchprop contains a database switching proposal
type Switchprop struct {
	Num   int    `json:"Numero"`
	SWUrl string `json:"Url"`
	Fnr   int    `json:"Fnr"`
}

// SwitchList contains list of mnemonic mapping
var SwitchList map[string]*Switchprop

// Config is the structure for all globals parameters
type Config struct {
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
var Conf Config

// GetConfig allows other module to get access to parameters in config file
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

func init() {
	conffyle := os.Args[1]
	fic, ficerr := ioutil.ReadFile(conffyle)
	if ficerr != nil {
		fmt.Printf("error ioutil : %v \n", ficerr)
	}
	_ = json.Unmarshal([]byte(fic), &Conf)
	switcherfyle := Conf.Switcherconf
	fichier, _ := ioutil.ReadFile(switcherfyle)
	_ = json.Unmarshal([]byte(fichier), &SwitchList)
	fmt.Printf("resume de l'init\n=================\n")
	fmt.Printf("le parametre reçu:%v\n", conffyle)
	fmt.Printf("le fichier param :%v\n", string(fic))
	fmt.Printf("le nom du fichier switch:%v\n", switcherfyle)
	fmt.Printf("le fichier switching:%v\n", string(fichier))
}

func result2struct(r *adabas.Response) []Carinfo {
	var cl []Carinfo
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
		cl = append(cl, voiture)
	}
	return cl
}

// Adaswitch create a adabas connection string based on received value
func Adaswitch(i interface{}) {
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
		Conf.Dburl = fmt.Sprintf("%v(adatcp://%v)", SwitchList[clef].Num, SwitchList[clef].SWUrl)
		Conf.ConnectString = fmt.Sprintf("%v,%v", Conf.Dburl, SwitchList[clef].Fnr)
		Conf.Fnr = uint32(SwitchList[clef].Fnr)
	}
	fmt.Printf("configuring Dburl to %v\n", Conf.Dburl)
	fmt.Printf("configuring Dbconn to %v\n", Conf.ConnectString)
}

// LoadMyJSONMap will load the maps to the DB
func LoadMyJSONMap(Myconn *adabas.Connection) {
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
	Adaswitch(Conf.Dbtouse)
	// creating the connection
	fullconn := "acj;map;config=[" + Conf.ConnectString + "]"
	fmt.Printf("Opening connection to :%v\n", fullconn)
	adaConnect, err := adabas.NewConnection(fullconn)
	if err != nil {
		fmt.Printf("NewConnection() error=%v\n", err)
	}
	LoadMyJSONMap(adaConnect)
	return adaConnect
}

// Carslist return the list of cars up to number given by limite
func Carslist(Myconn *adabas.Connection, limite uint64) CarList {
	// creating Read Request with MAP
	readRequest, cerr := Myconn.CreateMapReadRequest("VehicleMap")
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
	carlist.Vehicules = result2struct(result)
	return carlist
}

// CarsSearch return the list of cars for the specified vendors
func CarsSearch(Myconn *adabas.Connection, vv string, mv string, cv string, limite uint64) CarList {
	// creating Read Request with MAP
	readRequest, cerr := Myconn.CreateMapReadRequest("VehicleMap")
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
	result, err := readRequest.ReadLogicalWith(cherche)
	if err != nil {
		fmt.Printf("ReadLogicalWith() error=%v\n", err)
	}
	var carlist CarList
	carlist.Vehicules = result2struct(result)
	return carlist
}

//AddCar tries to create a car record in the file
func AddCar(Myconn *adabas.Connection, vendeur string, modele string, couleur string) UpdCar {
	// creating Store Request with MAP
	storeRequest, cerr := Myconn.CreateMapStoreRequest("VehicleMap")
	if cerr != nil {
		fmt.Printf("CreateMapStoreRequest() error=%v\n", cerr)
	}
	// Assigning query's fields
	err := storeRequest.StoreFields("Vendor,Model,Color")
	if err != nil {
		fmt.Printf("QueryFields() error=%v\n", err)
	}
	enreg := &Carrec{Vendor: vendeur, Model: modele, Color: couleur}
	sterr := storeRequest.StoreData(enreg)
	if sterr != nil {
		fmt.Printf("store enreg error=%v\n", sterr)
	}
	var result UpdCar
	result.Vehicule.Vendor = vendeur
	result.Vehicule.Color = couleur
	result.Vehicule.Model = modele
	result.Opcode = true
	transerr := storeRequest.EndTransaction()
	if transerr != nil {
		result.Opcode = false
		fmt.Printf("End transaction error=%v\n", transerr)
	}
	return result
}

// DelCar enables user to delete a car
func DelCar(Myconn *adabas.Connection, carisn uint64) UpdCar {
	deleteRequest, cerr := Myconn.CreateMapDeleteRequest("VehicleMap")
	if cerr != nil {
		fmt.Printf("CreateMapDeleteRequest() error=%v\n", cerr)
	}

	sterr := deleteRequest.Delete(adatypes.Isn(carisn))
	if sterr != nil {
		fmt.Printf("store enreg error=%v\n", sterr)
	}
	var result UpdCar
	result.Opcode = true
	transerr := deleteRequest.EndTransaction()
	if transerr != nil {
		result.Opcode = false
		fmt.Printf("End transaction error=%v\n", transerr)
	}
	return result
}

//UpdateCar  enables update
func UpdateCar(Myconn *adabas.Connection, isn uint64, vendeur string, modele string, couleur string) UpdCar {
	// creating Store Request with MAP
	storeRequest, cerr := Myconn.CreateMapStoreRequest("VehicleMap")
	if cerr != nil {
		fmt.Printf("CreateMapStoreRequest() error=%v\n", cerr)
	}
	// Assigning query's fields
	err := storeRequest.StoreFields("Isn,Vendor,Model,Color")
	if err != nil {
		fmt.Printf("QueryFields() error=%v\n", err)
	}
	enreg := &CarRecU{Isn: adatypes.Isn(isn), Vendor: vendeur, Model: modele, Color: couleur}
	sterr := storeRequest.StoreData(enreg)
	if sterr != nil {
		fmt.Printf("store enreg error=%v\n", sterr)
	}
	var result UpdCar
	result.Vehicule.Vendor = vendeur
	result.Vehicule.Color = couleur
	result.Vehicule.Model = modele
	result.Opcode = true
	transerr := storeRequest.EndTransaction()
	if transerr != nil {
		result.Opcode = false
		fmt.Printf("End transaction error=%v\n", transerr)
	}
	return result
}
