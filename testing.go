package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
)

//var afs_x64_path = "/tmp"

var con Configuration

// Configuration
type Configuration struct {
	AFSPATH      string
	AFS_SYN_PATH string
}

func readConfig() Configuration {
	file, _ := os.Open("conf.json")
	defer file.Close()
	decoder := json.NewDecoder(file)
	configuration := Configuration{}
	err := decoder.Decode(&configuration)
	if err != nil {
		fmt.Println("error:", err)
	}
	fmt.Println(configuration)
	return configuration
}

//

func main() {
	con = readConfig()
	fmt.Println("afs-x64 path:" + string(con.AFSPATH))
	fmt.Println("Server Start")
	r := gin.Default()
	route_register(r)
	r.Run()
}
func route_register(router *gin.Engine) {
	fileHandling := router.Group("/file")
	{
		fileHandling.POST("/upload", uploadFile)
		fileHandling.GET("/download/:afid", downloadFile)
	}
	info := router.Group("/info")
	{
		info.GET("/:afid", searchFile)
	}
}

func downloadFile(c *gin.Context) {

	//afid := c.Param("afid")
	var filename = strconv.Itoa(rand.Intn(9999999999)) + ".dat"
	var command = ";_f=download;afid=" + "1e000000000005d61421f459848221480915bef51fa4f147d0d6aa5d2d5be899f05cef27dc732360c4d853aff3b616c2682beedb1f82ddf90811a7cb97bd0d08" + ";local_file=" + filename + ";"
	fmt.Println("Run command: " + string(con.AFSPATH) + command)
	runBashComamnd(command)

	dat, err := ioutil.ReadFile(filename)
	if err != nil {
		return
	}
	//content := c.Query("content")
	//content = dat + content
	c.Writer.WriteHeader(http.StatusOK)
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", "application/text/plain")
	c.Header("Accept-Length", fmt.Sprintf("%d", len(dat)))
	c.Writer.Write(dat)
}

func f(wg *sync.WaitGroup, afid string, i int, isExist []bool, c *gin.Context) {
	defer wg.Done()
	print(isExists(afid, c))
	//isExist[i] = true
}

func searchFile(c *gin.Context) {
	afid := c.Param("afid")
	var addressList = readrNodeList()
	var isExist []bool
	var result gin.H

	var wg sync.WaitGroup
	// Multi Thread
	for i := 0; i < len(addressList); i++ {
		wg.Add(1)
		go f(&wg, afid, i, isExist, c)
	}
	wg.Wait()
	result = gin.H{
		"result": isExist,
		"err":    nil}
	c.JSON(http.StatusOK, result)
}

func uploadFile(c *gin.Context) {

	var result gin.H
	var commandResult string

	file, err := c.FormFile("file")
	if err != nil {
		result = gin.H{
			"result": nil,
			"err":    "Get form err"}
		c.JSON(http.StatusBadRequest, result)
		return
	}

	if err := c.SaveUploadedFile(file, file.Filename); err != nil {
		result = gin.H{
			"result": nil,
			"err":    "upload file err"}
		c.JSON(http.StatusBadRequest, result)
		return
	}
	fmt.Println("Upload file xx" + string(file.Filename))
	var command = ";_f=upload;local_file=" + string(file.Filename) + ";"
	fmt.Println("Run command: " + string(con.AFSPATH) + command)
	commandResult = runBashComamnd(command)
	result = gin.H{
		"result": commandResult,
		"err":    nil}
	c.JSON(http.StatusOK, result)
}

func isExists(afid string, c *gin.Context) string {

	var command = ";_f=query;afid=" + afid + ";"
	var commandResult = runBashComamnd(command)
	return commandResult
}

func readrNodeList() []string {

	var addressList []string
	file, err := os.Open(con.AFS_SYN_PATH)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		addressList = append(addressList, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return addressList
}

func runBashComamnd(command string) string {

	var cmd *exec.Cmd
	var commandResult []byte
	var err error
	// string(con.AFSPATH),command
	cmd = exec.Command(string(con.AFSPATH), command)
	if commandResult, err = cmd.CombinedOutput(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println(string(commandResult))
	return string(commandResult)
}
