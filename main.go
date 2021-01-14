package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	"github.com/elastic/go-elasticsearch"
	"github.com/elastic/go-elasticsearch/esapi"
)

// AlertStruct is a struct defining alert structure that is going to be saved to ES
type AlertStruct struct {
	TimeCreated int64
	Name        string
	Software    string `json:"software"`
	Structure   string `json:"structure"`
}

// Method for converting AlertStruct to json
func (a AlertStruct) jsonStruct() string {
	// Create struct instance of the Elasticsearch fields struct object
	docStruct := &AlertStruct{
		TimeCreated: a.TimeCreated,
		Name:        a.Name,
		Software:    a.Software,
		Structure:   a.Structure,
	}
	// Marshal the struct to JSON and check for errors
	b, err := json.Marshal(docStruct)
	if err != nil {
		fmt.Println("json.Marshal ERROR:", err)
		return string(err.Error())
	}
	return string(b)
}

// httpClientResponse is a sturct defining API response structure
type httpClientResponseFormat struct {
	Status        string `json:"status"`
	Message       string `json:"message"`
	AppStatusCode int    `json:"appStatusCode"`
}

func main() {
	// Instantiate new mux router and create first route. Configure logging with github.com/gorilla/handlers/CombinedLoggingHandler
	router := mux.NewRouter()
	router.Handle("/alert/structure/{alertStructureName}", handlers.CombinedLoggingHandler(os.Stdout, http.HandlerFunc(addAlertStructure)))

	http.ListenAndServe(":3000", router)
}

// Handler for /alert/structure/{alertStructureName} route
func addAlertStructure(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	alertName := mux.Vars(r)["alertStructureName"]

	// Switch for different HTTP Methods on this endpoint
	switch r.Method {
	case "GET":

	case "POST":
		timesamp := time.Now()

		esDocExists, esDocExistsErr := checkExistingAlertStructure(alertName)

		if esDocExistsErr != nil {
			//Error is beeing logged at function level
			//Craft response payload for client
			clientRespose := httpClientResponseFormat{"failure", "internal_error", 5001}
			//Send response status code and payload
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(clientRespose)
		} else if esDocExists == true {
			//Craft response payload for client
			clientRespose := httpClientResponseFormat{"failure", "alet_structure_exists", 4001}
			//Send response status code and payload
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(clientRespose)
		} else if esDocExists == false {
			//Create new alert structure object with name from path and creation timestamp
			newAlertStructure := AlertStruct{Name: alertName, TimeCreated: timesamp.Unix()}
			//Append user data to new alert structure object
			json.NewDecoder(r.Body).Decode(&newAlertStructure)

			//Try to save document to ES
			esIndexedDoc, esIndexingErr := saveToElasticsearch(newAlertStructure)

			if esIndexedDoc != "" && esIndexingErr != nil {
				//Log indexing error to console
				log.Printf(esIndexingErr.Error())
				//Craft response payload for client
				clientRespose := httpClientResponseFormat{"failure", "internal_error", 5002}
				//Send response status code and payload
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(clientRespose)
			} else {
				//Return status created and created document from ES
				w.WriteHeader(http.StatusCreated)
				clientRespose := httpClientResponseFormat{"success", esIndexedDoc, 2000}
				json.NewEncoder(w).Encode(clientRespose)
			}

		}
	case "PUT":
		json.NewEncoder(w).Encode("method is " + r.Method)
	case "DELETE":
		json.NewEncoder(w).Encode("method is " + r.Method)
	}

}

func saveToElasticsearch(doc AlertStruct) (string, error) {
	// Create a context object for the API calls
	ctx := context.Background()

	cfg := elasticsearch.Config{
		Addresses: []string{
			"http://localhost:9200",
		},
	}
	esClient, esClientErr := elasticsearch.NewClient(cfg)
	if esClientErr != nil {
		fmt.Println("Elasticsearch connection error:", esClientErr)
	}

	// Instantiate a request object
	indexReq := esapi.IndexRequest{
		Index: "2alerts",
		Body:  strings.NewReader(doc.jsonStruct()), //Call method and convert doc type AlertStruct to JSON
	}

	// Return an API response object from request
	indexRes, indexResErr := indexReq.Do(ctx, esClient)
	if indexResErr != nil {
		return "", indexResErr
	}
	defer indexRes.Body.Close()

	if indexRes.IsError() {
		log.Printf("%s ERROR indexing document", indexRes.Status())
	} else {
		// Deserialize the response into a map.
		var r map[string]interface{}
		if err := json.NewDecoder(indexRes.Body).Decode(&r); err != nil {
			log.Printf("Error parsing the response body: %s", err)
		} else {
			return fmt.Sprintf("%v", r["_id"]), nil
		}
	}
	return "", nil
}

func checkExistingAlertStructure(alertStructureName string) (bool, error) {
	var (
		r map[string]interface{}
	)

	// Create a context object for the API calls
	//ctx := context.Background()

	cfg := elasticsearch.Config{
		Addresses: []string{
			"http://localhost:9200",
		},
	}
	esClient, esClientErr := elasticsearch.NewClient(cfg)
	if esClientErr != nil {
		log.Printf("Elasticsearch connection error: %s", esClientErr)
		return false, esClientErr
	}
	var buf bytes.Buffer
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"match": map[string]interface{}{
				"Name": alertStructureName,
			},
		},
	}
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		log.Printf("Error encoding query: %s", err)
		return false, err
	}

	res, err := esClient.Info()
	if err != nil {
		log.Printf("Error getting response: %s", err)
		return false, err
	}
	// Perform the search request.
	res, err = esClient.Search(
		esClient.Search.WithContext(context.Background()),
		esClient.Search.WithIndex("2alerts"),
		esClient.Search.WithBody(&buf),
		esClient.Search.WithTrackTotalHits(true),
		esClient.Search.WithPretty(),
	)
	if err != nil {
		log.Printf("Error getting response: %s", err)
		return false, err
	}
	defer res.Body.Close()

	if res.IsError() {
		var e map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			log.Printf("Error parsing the response body: %s", err)
		} else {
			// Print the response status and error information.
			log.Printf("[%s] %s: %s",
				res.Status(),
				e["error"].(map[string]interface{})["type"],
				e["error"].(map[string]interface{})["reason"],
			)
		}
		return false, err
	}

	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		log.Printf("Error parsing the response body: %s", err)
		return false, err
	}
	/* Print the response status, number of results, and request duration.
	log.Printf(
		"[%s] %d hits; took: %dms",
		res.Status(),
		int(r["hits"].(map[string]interface{})["total"].(map[string]interface{})["value"].(float64)),
		int(r["took"].(float64)),
	)
	*/
	//Check if map r under key hists and value total (which is also a map), is eq to 0 -> if es query returned any results
	if int(r["hits"].(map[string]interface{})["total"].(map[string]interface{})["value"].(float64)) == 0 {
		return false, nil
	}
	//Return true by default which means above if failed and there are alert structures with same name
	return true, nil
}
