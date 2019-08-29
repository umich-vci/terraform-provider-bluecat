package bluecat

import (
	"fmt"
	"log"
	"net/http"

	"github.com/fiorix/wsdl2go/soap"
	"github.com/umich-vci/golang-bluecat"
)

// Config holds the provider configuration
type Config struct {
	Username        string
	Password        string
	BlueCatEndpoint string
	SessionCookies  []*http.Cookie
}

// Client returns a new client for accessing BlueCat Address Manager
func (c *Config) Client() (bam.ProteusAPI, error) {
	//var response *http.Response
	cli := soap.Client{
		URL:       "https://" + c.BlueCatEndpoint + "/Services/API?wsdl",
		Namespace: bam.Namespace,
		Pre:       c.setBlueCatAuthToken,
		Post:      c.getBlueCatAuthToken,
	}
	soapService := bam.NewProteusAPI(&cli)
	log.Printf("[INFO] BlueCat URL is: %s", cli.URL)
	err := soapService.Login(c.Username, c.Password)
	if err != nil {
		return nil, fmt.Errorf("Login error: %s", err)
	}
	log.Printf("[INFO] BlueCat Login was successful")

	return soapService, nil
}

func (c *Config) setBlueCatAuthToken(request *http.Request) {
	for i := range c.SessionCookies {
		request.AddCookie(c.SessionCookies[i])
	}
}

func (c *Config) getBlueCatAuthToken(response *http.Response) {
	(*c).SessionCookies = append((*c).SessionCookies, response.Cookies()...)
}
