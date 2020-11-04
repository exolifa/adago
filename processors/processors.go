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
		"payload": mycarlist}, "carform.html")

}

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
		msg := dbproc.DelCar(Myconn, uint64(i64))
		mycarlist := dbproc.CarsSearch(Myconn, vendel, "", "", 0)
		mycarlist.Message = msg
		fmt.Printf("data received from adabas:%v\n", mycarlist)
		render(c, gin.H{
			"title":   "List of cars",
			"payload": mycarlist}, "carform.html")

	case "CREATE":
		vencr := c.PostForm("vencr")
		colcr := c.PostForm("colcr")
		modcr := c.PostForm("modcr")
		msg := dbproc.AddCar(Myconn, vencr, modcr, colcr)
		mycarlist := dbproc.CarsSearch(Myconn, vencr, "", "", 0)
		mycarlist.Message = msg
		fmt.Printf("data received from adabas:%v\n", mycarlist)
		render(c, gin.H{
			"title":   "List of cars",
			"payload": mycarlist}, "carform.html")
	case "UPDATE":
		venup := c.PostForm("vencr")
		colup := c.PostForm("colcr")
		modup := c.PostForm("modcr")
		isnup := c.PostForm("isn")
		i64, _ := strconv.ParseInt(isnup, 10, 64)
		msg := dbproc.UpdateCar(Myconn, uint64(i64), venup, modup, colup)
		mycarlist := dbproc.CarsSearch(Myconn, venup, modup, colup, 0)
		mycarlist.Message = msg
		fmt.Printf("data received from adabas:%v\n", mycarlist)
		render(c, gin.H{
			"title":   "List of cars",
			"payload": mycarlist}, "carform.html")

	case "SELECT":
		vensel := c.PostForm("vensel")
		modsel := c.PostForm("modsel")
		colsel := c.PostForm("colsel")
		fmt.Printf("searching for Vendor: %v Model %v and Color:%v\n", vensel, modsel, colsel)
		mycarlist := dbproc.CarsSearch(Myconn, vensel, modsel, colsel, 0)
		fmt.Printf("data received from adabas:%v\n", mycarlist)
		render(c, gin.H{
			"title":   "List of cars",
			"payload": mycarlist}, "carform.html")
	default:
		mycarlist := dbproc.Carslist(Myconn, 0)
		render(c, gin.H{
			"title":   "List of cars",
			"payload": mycarlist}, "carform.html")

	}
}
