# Demo Go web application using ADABAS

## Pre-requisites

This small application assumed the presence of an ADABAS server housing the demo database provided by Software AG. The simpliest way to have the required DB server is to install the community edition in container (ADABAS-CE container) that can be downloaded free of charge on Docker-Hub.
When the container is started, if it is desirable to keep the data, then the docker should be started with permanent storage assigned (ref. to Software AG documentation)
The demo application uses maps to access data-fields, therefore it is required to create in the demo database a space (File) to store the maps (this is not part of the demo datase) . Therefore an proper FDT and FDU are needed. 

# Installation 

If the windows exe file is downloaded , after the parameters files are adjusted , it is ready. 
The demo is provided as a Go project using modules. 

## Main parameters file 

{
"db_to_use": "Production",
"switcher_config": "<path to switcher config file ><switcher config file>",
"maps_file": "<path to map  file>/vehicules.json",
"templates": "<path to template directory>*",
"port": "8999"
}
- The db_to_use should be one of the mnemonic/acronym described in the switcher config. 
- Port must be an available port on the machine where the demo is installed
- On windows system, path are of the form c:\\dir1\\subdir2\\...etc . With the double \. 

## Switcher parameters file 
- 
This is just for convenience purpose to be able to adapt the DB parameters to the local implementation without having to modify the program (and therefore require a usuable Go environment.)
{
"Production":{
      "Numero" : 12 ,
      "Url": "localhost:60001",
      "Fnr": 4}
}
"Production" : is the mnemonic and can be whatever as fare the same mnemo is used in the Main parameters file. 
"Numéro" : is the number of the Adabas DB .Currently the demo db has number 12.
"Url' : this is the URL to the adabas docker container
"Fnr": this is the file number where the map are going to be stored. It must be the File number used in the creation of the Map  file (see : Pre-requisites)

## Vehicules Map file 
This file is the minimalist declaration of the mapping (IE: define correspondance between a long "friendly name" and the  Adabas internal field name (typically on 2 alphanumeric).
Modification to this file will require modification of the program. 
{"Maps":[
      {"Name":"VehicleMap",
       "Data":{
                 "Target":"12(adatcp://localhost:60001)","File":12},
                 "LastModifified":"2020\\10\\06 20:11:26",
                 "Fields":[
                        {"LongName":"Vendor","ShortName":"AD","ContentType":"","Charset":"US-ASCII","File":0,"FormatType":"A","FormatLength":-1,"FieldType":"ALPHA"},
                       {"LongName":"Model","ShortName":"AE","ContentType":"","Charset":"US-ASCII","File":0,"FormatType":"A","FormatLength":-1,"FieldType":"ALPHA"},
                     {"LongName":"Color","ShortName":"AF","ContentType":"","Charset":"US-ASCII","File":0,"FormatType":"A","FormatLength":-1,"FieldType":"ALPHA"}
               ]
      }]
}
To modify, please refer to the map section of the ADABAS-GO-API and modify the related access in the Go modules.

## Run of the demo 

On windows , using a command prompt: 
c:> adago.exe <path to Main Parameters File> 
The program will initiate connection to ADABAS server and setup the  webserver (based on gin-gonic) on the port configured in the Main Parameters. 
You can rename the current file by clicking the file name in the navigation bar or by clicking the **Rename** button in the file explorer.

`expected output
==============
resume de l'init
le parametre reçu:<file given as argument to the exe>
le fichier param :{
    ................
    Main Parameters File content
    ................
}
le nom du fichier <switcher file full path>
le fichier switching:{
	................
	Switche config file content
	................
}
configuring Dburl to 12(adatcp://localhost:60001)
configuring Dbconn to 12(adatcp://localhost:60001),4
Opening connection to :acj;map;config=[12(adatcp://localhost:60001),4]
loading map from dbproc/vehicules.json into :12(adatcp://localhost:60001)
Loading ....<Map file path>
ADDING MAP TO GLOBAL REPO
        ADDING FNR 4 AS GLOBAL REPO
        STORE MAP TO REPO
>Eventually if the map is already loaded an error message is displayed:
store map error=ADAGE62000: Unique descriptor already present (rsp=98,subrsp=0,dbid=12(adatcp://localhost:60001),file=4)

ADDING MAP TO CACHE
[GIN-debug] [WARNING] Creating an Engine instance with the Logger and Recovery middleware already attached.

[GIN-debug] [WARNING] Running in "debug" mode. Switch to "release" mode in production.
 - using env:   export GIN_MODE=release
 - using code:  gin.SetMode(gin.ReleaseMode)

[GIN-debug] Loaded HTML Templates (8):
        - carform.html
        - carlist.html
        - carupd.html
        - footer.html
        - header.html
        - menu.html
        - carcreate.html

[GIN-debug] GET    /                         --> adago.net/adago/processors.Carsliste (3 handlers)
[GIN-debug] POST   /formcars        --> adago.net/adago/processors.FormCars (3 handlers)
**[GIN-debug] Listening and serving HTTP on :8999**`

When the init phase is completed , in a browser window , go to "localhost:8999" this sould bring the main window with all the car listed. 

# Important remark

This project was made as an exploration to the GO-API functionalities... not as a tutorial of best practices neither on Go programming nor on Adabas usage.
The purpose was just to demonstrate the ease of integrating ADABAS data in Go program.

## About the author

I stopped writing code more than 2 decades ago. Today I still produce some "code" for my hobby mainly on ESP/Arduino platforms. 
I discovered ADABAS because ,as part of my management assigments, I make with our former Mainframe team the long trip to move the existing Cobol/Natural/Adabas applications to Linux and now to Kubernetes.
Our main Enterprise Architect introduced me to Golang as part of our regular marketplace's survey. I got a "click": could be because some aspects remind me of my coding experience in Algol ,long but also because the learning curve was easy. I like Golang for its powerful simplicity...  
