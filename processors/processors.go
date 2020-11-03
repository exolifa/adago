package processors

import (
	"fmt"
	"net/http"
	"strings"

	"strconv"

	"adago.net/adago/dbproc"

	"github.com/gin-gonic/gin"
)

// Myconn est le contenu de la connexion ADABAS
var Myconn = dbproc.Adabasinit()

// Showmenu send back the index.html page
func Showmenu(c *gin.Context) {
	fmt.Printf("Request to send menu")
	// Call the HTML method of the Context to render a template
	c.HTML(
		// Set the HTTP status to 200 (OK)
		http.StatusOK,
		// Use the index.html template
		"carform.html",
		// Pass the data that the page uses (in this case, 'title')
		gin.H{
			"title": "Home Page",
		},
	)

}

// Render to handle all types of request (html ,json,xml

func render(c *gin.Context, data gin.H, templateName string) {
	switch c.Request.Header.Get("Accept") {
	case "application/json":
		// Respond with JSON
		c.JSON(http.StatusOK, data["payload"])
	case "application/xml":
		// Respond with XML
		c.XML(http.StatusOK, data["payload"])
	default:
		// Respond with HTML
		c.HTML(http.StatusOK, templateName, data)
	}
}

// Carsliste return full list of cars sorted by vendor
func Carsliste(c *gin.Context) {
	defer Myconn.Close()
	err := Myconn.Open()
	if err != nil {
		fmt.Printf("Open() error=%v\n", err)
		return
	}
	mycarlist := dbproc.Carslist(Myconn, 0)
	render(c, gin.H{
		"title":   "List of cars",
		"payload": mycarlist.Vehicules}, "carform.html")

}

// CarsParam return full list of cars sorted by vendor
/*func CarsParam(c *gin.Context) {
	defer Myconn.Close()
	err := Myconn.Open()
	if err != nil {
		fmt.Printf("Open() error=%v\n", err)
		return
	}
	clef := c.Param("clef")
	valeur := c.Param("valeur")
	mycarlist := dbproc.CarsSearch(Myconn, clef, valeur, "", "", 0)
	render(c, gin.H{
		"title":   "List of cars",
		"payload": mycarlist.Vehicules}, "carform.html")
}*/

// FormCars return full list of cars sorted by vendor
func FormCars(c *gin.Context) {
	defer Myconn.Close()
	err := Myconn.Open()
	if err != nil {
		fmt.Printf("Open() error=%v\n", err)
		return
	}
	action := c.PostForm("oper")
	fmt.Printf("in formcars with action=%v\n", action)
	switch action {
	case "ADD":
		render(c, gin.H{
			"title": "Add New car",
		}, "carcreate.html")
	case "UPD":
		cible := c.PostForm("selection")
		s := strings.Split(cible, "-")
		isnupd, venupd, modupd, colupd := s[0], s[1], s[2], s[3]
		i64, _ := strconv.ParseInt(isnupd, 10, 64)
		var carupd []dbproc.Carinfo
		var voiture dbproc.Carinfo
		voiture.Isn = uint64(i64)
		voiture.Vendor = venupd
		voiture.Color = colupd
		voiture.Model = modupd
		carupd = append(carupd, voiture)
		render(c, gin.H{
			"title":   "Update car",
			"payload": carupd}, "carupd.html")

	case "DELETE":
		cible := c.PostForm("selection")
		s := strings.Split(cible, "-")
		isndel, vendel := s[0], s[1]
		i64, _ := strconv.ParseInt(isndel, 10, 64)
		_ = dbproc.DelCar(Myconn, uint64(i64))
		mycarlist := dbproc.CarsSearch(Myconn, vendel, "", "", 0)
		fmt.Printf("data received from adabas:%v\n", mycarlist)
		render(c, gin.H{
			"title":   "List of cars",
			"payload": mycarlist.Vehicules}, "carform.html")

	case "CREATE":
		vencr := c.PostForm("vencr")
		colcr := c.PostForm("colcr")
		modcr := c.PostForm("modcr")
		_ = dbproc.AddCar(Myconn, vencr, modcr, colcr)
		mycarlist := dbproc.CarsSearch(Myconn, vencr, "", "", 0)
		fmt.Printf("data received from adabas:%v\n", mycarlist)
		render(c, gin.H{
			"title":   "List of cars",
			"payload": mycarlist.Vehicules}, "carform.html")
	case "UPDATE":
		venup := c.PostForm("vencr")
		colup := c.PostForm("colcr")
		modup := c.PostForm("modcr")
		isnup := c.PostForm("isn")
		i64, _ := strconv.ParseInt(isnup, 10, 64)
		_ = dbproc.UpdateCar(Myconn, uint64(i64), venup, modup, colup)
		mycarlist := dbproc.CarsSearch(Myconn, venup, modup, colup, 0)
		fmt.Printf("data received from adabas:%v\n", mycarlist)
		render(c, gin.H{
			"title":   "List of cars",
			"payload": mycarlist.Vehicules}, "carform.html")

	case "SELECT":
		vensel := c.PostForm("vensel")
		modsel := c.PostForm("modsel")
		colsel := c.PostForm("colsel")
		fmt.Printf("searching for Vendor: %v Model %v and Color:%v\n", vensel, modsel, colsel)
		mycarlist := dbproc.CarsSearch(Myconn, vensel, modsel, colsel, 0)
		fmt.Printf("data received from adabas:%v\n", mycarlist)
		render(c, gin.H{
			"title":   "List of cars",
			"payload": mycarlist.Vehicules}, "carform.html")
	default:
		mycarlist := dbproc.Carslist(Myconn, 0)
		render(c, gin.H{
			"title":   "List of cars",
			"payload": mycarlist.Vehicules}, "carform.html")

	}
}

// CarAdd create a new car in the dabaser
func CarAdd(c *gin.Context) {
	defer Myconn.Close()
	err := Myconn.Open()
	if err != nil {
		fmt.Printf("Open() error=%v\n", err)
		return
	}
	vendeur := c.Param("vendeur")
	modele := c.Param("modele")
	couleur := c.Param("couleur")
	updrsp := dbproc.AddCar(Myconn, vendeur, modele, couleur)
	fmt.Printf("update response:%v\n", updrsp)
	render(c, gin.H{
		"title":   "Creation of car",
		"payload": updrsp}, "carcreate.html")
}
