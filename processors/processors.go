package processors

import (
	"fmt"
	"net/http"
	"strings"

	"strconv"

	"adago.net/adago/dbproc"

	"github.com/gin-gonic/gin"
)

// Myconn is the adabas connection
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
	//First 2 cases to go to dedicated screens
	//ADD returns the empty screen to create a new vehicule
	case "ADD":
		render(c, gin.H{
			"title": "Add New car",
		}, "carcreate.html")
		//UPD returns a prefilled screen to modify data of a vehicule
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
		// Now the real actions
		// DELETE allows deletion of the car selected in the main screen
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
		// From the ADDITION screen this will ADD the car to ADABAS
		// The return screen show the list of cars from the vendor
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
		// From the UPD screen this modify the record (based on the ISN /Reference number of the car)
		// It returns the modified record
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
		// Based on the combination of the 3 criteria it,
		// SELECT returns the car(s) complying with the combined criteria
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
		// In case something goes wrong, this send back to the main screen
		//showing all cars in the database
	default:
		mycarlist := dbproc.Carslist(Myconn, 0)
		render(c, gin.H{
			"title":   "List of cars",
			"payload": mycarlist}, "carform.html")

	}
}
