package routing

import (
	"adago.net/adago/dbproc"
	"adago.net/adago/processors"

	"github.com/gin-gonic/gin"
)

// SetRoutes defines de different routes for gin service
func SetRoutes() *gin.Engine {

	r := gin.Default()
	templatedir := dbproc.GetConfig("Templatesdir")
	r.LoadHTMLGlob(templatedir)
	r.GET("/", processors.Carsliste)
	r.GET("/carlist", processors.Carsliste)
	//	r.GET("/carparam/:clef/:valeur", processors.CarsParam)
	r.GET("/caradd/:vendeur/:modele/:couleur", processors.CarAdd)
	r.POST("/formcars", processors.FormCars)
	return r
}
